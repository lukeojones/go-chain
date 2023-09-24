package main

import (
	"crypto/sha256"
	"math"
)

type MerkleNode struct {
	Left  *MerkleNode
	Right *MerkleNode
	Data  []byte
}

type MerkleTree struct {
	Root *MerkleNode
}

func NewMerkleNode(left, right *MerkleNode, data []byte) *MerkleNode {
	node := MerkleNode{}
	var hash [32]byte
	if left == nil && right == nil {
		hash = sha256.Sum256(data)
	} else {
		hashChildren := append(left.Data, right.Data...)
		hash = sha256.Sum256(hashChildren)
	}
	node.Data = hash[:]
	node.Left = left
	node.Right = right

	return &node
}

// NewMerkleTree creates a merkle tree from a serialized list of transactions
func NewMerkleTree(data [][]byte) *MerkleTree {
	// Copy the last element to ensure there are an even number of nodes
	if len(data)%2 != 0 {
		data = append(data, data[len(data)-1])
	}

	var nodes []MerkleNode

	// Build a leaf node for every piece of data
	for _, datum := range data {
		node := NewMerkleNode(nil, nil, datum)
		nodes = append(nodes, *node)
	}

	// Iteratively build up the levels of the tree
	// (each level has half the length of the level below it)
	nLevels := int(math.Ceil(math.Log(float64(len(data)))))
	for i := 0; i < nLevels; i++ {
		var level []MerkleNode
		for j := 0; j < len(nodes); j += 2 {
			parent := NewMerkleNode(&nodes[j], &nodes[j+1], nil)
			level = append(level, *parent)
		}
		nodes = level
	}

	// There should only be one node here anyway.
	return &MerkleTree{Root: &nodes[0]}
}
