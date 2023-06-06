// Copyright 2020 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package trie

import (
	"bytes"
	"fmt"
	"github.com/ethereum/go-ethereum/rlp"

	"github.com/ethereum/go-ethereum/common"
)

// leaf represents a trie leaf node
type leaf struct {
	blob   []byte      // raw blob of leaf
	parent common.Hash // the hash of parent node
}

// committer is the tool used for the trie Commit operation. The committer will
// capture all dirty nodes during the commit process and keep them cached in
// insertion order.
type committer struct {
	nodes       *NodeSet
	collectLeaf bool
	saveNode    map[string][]byte
}

// newCommitter creates a new committer or picks one from the pool.
func newCommitter(owner common.Hash, collectLeaf bool) *committer {
	return &committer{
		nodes:       NewNodeSet(owner),
		collectLeaf: collectLeaf,
		saveNode:    map[string][]byte{},
	}
}

func (c *committer) SetDelta(delta []*NodeDelta) {
	for _, d := range delta {
		c.saveNode[d.Key] = d.Val
	}
}

func (c *committer) GetDelta() []*NodeDelta {
	delta := make([]*NodeDelta, 0, len(c.saveNode))
	for k, v := range c.saveNode {
		delta = append(delta, &NodeDelta{k, v})
	}
	return delta
}

// Commit collapses a node down into a hash node and inserts it into the database
func (c *committer) Commit(n node) (hashNode, *NodeSet, error) {
	var h node
	var err error
	if len(c.saveNode) > 0 {
		rootHash := c.saveNode["root"]
		h, err = c.commitWithDelta(nil, rootHash)
	} else {
		switch cn := n.(type) {
		case *shortNode:
			c.saveNode["root"] = cn.flags.hash
		case *fullNode:
			c.saveNode["root"] = cn.flags.hash
		case hashNode:
			c.saveNode["root"] = cn
		default:
			// nil, valuenode shouldn't be committed
			panic(fmt.Sprintf("%T: invalid node: %v", n, n))
		}
		h, err = c.commit(nil, n)
	}
	if err != nil {
		return nil, nil, err
	}
	return h.(hashNode), c.nodes, nil
}

func (c *committer) commitWithDelta(path, nodeHash []byte) (node, error) {
	if c.saveNode[string(nodeHash)] == nil {
		var hn hashNode = nodeHash
		return hn, nil
	}
	n := mustDecodeNode(nodeHash, c.saveNode[string(nodeHash)])
	// Commit children, then parent, and remove remove the dirty flag.
	switch cn := n.(type) {
	case *shortNode:
		// If the child is fullnode, recursively commit.
		// Otherwise it can only be hashNode or valueNode.
		if h, ok := cn.Val.(*hashNode); ok {
			_, err := c.commitWithDelta(append(path, cn.Key...), *h)
			if err != nil {
				return nil, err
			}
		} else if h, ok := cn.Val.(hashNode); ok {
			_, err := c.commitWithDelta(append(path, cn.Key...), h)
			if err != nil {
				return nil, err
			}
		}
		// The key needs to be copied, since we're delivering it to database
		cn.Key = hexToCompact(cn.Key)
		hashedNode := c.store(path, cn)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn, nil
		}
		return cn, nil
	case *fullNode:
		err := c.commitChildrenWithDelta(path, cn)
		if err != nil {
			return nil, err
		}

		hashedNode := c.store(path, cn)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn, nil
		}
		return cn, nil
	case hashNode:
		return cn, nil
	default:
		// nil, valuenode shouldn't be committed
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

// commit collapses a node down into a hash node and inserts it into the database
func (c *committer) commit(path []byte, n node) (node, error) {
	// if this path is clean, use available cached data
	hash, dirty := n.cache()
	if hash != nil && !dirty {
		return hash, nil
	}
	// Commit children, then parent, and remove the dirty flag.
	switch cn := n.(type) {
	case *shortNode:
		// Commit child
		collapsed := cn.copy()

		// If the child is fullNode, recursively commit,
		// otherwise it can only be hashNode or valueNode.
		if _, ok := cn.Val.(*fullNode); ok {
			childV, err := c.commit(append(path, cn.Key...), cn.Val)
			if err != nil {
				return nil, err
			}
			collapsed.Val = childV
		}
		// The key needs to be copied, since we're delivering it to database
		collapsed.Key = hexToCompact(cn.Key)

		// for dds producer
		var w bytes.Buffer
		if err := rlp.Encode(&w, collapsed); err != nil {
			panic("encode error: " + err.Error())
		}
		c.saveNode[string(collapsed.flags.hash)] = w.Bytes()

		hashedNode := c.store(path, collapsed)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn, nil
		}
		return collapsed, nil
	case *fullNode:
		hashedKids, err := c.commitChildren(path, cn)
		if err != nil {
			return nil, err
		}
		collapsed := cn.copy()
		collapsed.Children = hashedKids

		// for dds producer
		var w bytes.Buffer
		if err := collapsed.EncodeRLP(&w); err != nil {
			panic("encode error: " + err.Error())
		}
		c.saveNode[string(collapsed.flags.hash)] = w.Bytes()

		hashedNode := c.store(path, collapsed)
		if hn, ok := hashedNode.(hashNode); ok {
			return hn, nil
		}
		return collapsed, nil
	case hashNode:
		return cn, nil
	default:
		// nil, valuenode shouldn't be committed
		panic(fmt.Sprintf("%T: invalid node: %v", n, n))
	}
}

func (c *committer) commitChildrenWithDelta(path []byte, n *fullNode) error {
	for i := 0; i < 16; i++ {
		child := n.Children[i]
		if child == nil {
			continue
		}
		// If it's the hashed child, save the hash value directly.
		// Note: it's impossible that the child in range [0, 15]
		// is a valuenode.
		if hn, ok := child.(*hashNode); ok {
			_, err := c.commitWithDelta(append(path, byte(i)), *hn)
			if err != nil {
				return err
			}
		} else if hn, ok := child.(hashNode); ok {
			_, err := c.commitWithDelta(append(path, byte(i)), hn)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// commitChildren commits the children of the given fullnode
func (c *committer) commitChildren(path []byte, n *fullNode) ([17]node, error) {
	var children [17]node
	for i := 0; i < 16; i++ {
		child := n.Children[i]
		if child == nil {
			continue
		}
		// If it's the hashed child, save the hash value directly.
		// Note: it's impossible that the child in range [0, 15]
		// is a valueNode.
		if hn, ok := child.(hashNode); ok {
			children[i] = hn
			continue
		}
		// Commit the child recursively and store the "hashed" value.
		// Note the returned node can be some embedded nodes, so it's
		// possible the type is not hashNode.
		hashed, err := c.commit(append(path, byte(i)), child)
		if err != nil {
			return children, err
		}
		children[i] = hashed
	}
	// For the 17th child, it's possible the type is valuenode.
	if n.Children[16] != nil {
		children[16] = n.Children[16]
	}
	return children, nil
}

// store hashes the node n and if we have a storage layer specified, it writes
// the key/value pair to it and tracks any node->child references as well as any
// node->external trie references.
func (c *committer) store(path []byte, n node) node {
	// Larger nodes are replaced by their hash and stored in the database.
	var hash, _ = n.cache()

	// This was not generated - must be a small node stored in the parent.
	// In theory, we should check if the node is leaf here (embedded node
	// usually is leaf node). But small value(less than 32bytes) is not
	// our target(leaves in account trie only).
	if hash == nil {
		return n
	}
	// We have the hash already, estimate the RLP encoding-size of the node.
	// The size is used for mem tracking, does not need to be exact
	var (
		size  = estimateSize(n)
		nhash = common.BytesToHash(hash)
		mnode = &memoryNode{
			hash: nhash,
			node: simplifyNode(n),
			size: uint16(size),
		}
	)
	// Collect the dirty node to nodeset for return.
	c.nodes.add(string(path), mnode)

	// Collect the corresponding leaf node if it's required. We don't check
	// full node since it's impossible to store value in fullNode. The key
	// length of leaves should be exactly same.
	if c.collectLeaf {
		if sn, ok := n.(*shortNode); ok {
			if val, ok := sn.Val.(valueNode); ok {
				c.nodes.addLeaf(&leaf{blob: val, parent: nhash})
			}
		}
	}
	return hash
}

// estimateSize estimates the size of an rlp-encoded node, without actually
// rlp-encoding it (zero allocs). This method has been experimentally tried, and with a trie
// with 1000 leaves, the only errors above 1% are on small shortnodes, where this
// method overestimates by 2 or 3 bytes (e.g. 37 instead of 35)
func estimateSize(n node) int {
	switch n := n.(type) {
	case *shortNode:
		// A short node contains a compacted key, and a value.
		return 3 + len(n.Key) + estimateSize(n.Val)
	case *fullNode:
		// A full node contains up to 16 hashes (some nils), and a key
		s := 3
		for i := 0; i < 16; i++ {
			if child := n.Children[i]; child != nil {
				s += estimateSize(child)
			} else {
				s++
			}
		}
		return s
	case valueNode:
		return 1 + len(n)
	case hashNode:
		return 1 + len(n)
	default:
		panic(fmt.Sprintf("node type %T", n))
	}
}
