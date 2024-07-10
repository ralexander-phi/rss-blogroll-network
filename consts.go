package main

// TODO was int64
type NodeType = int

const (
	NODE_TYPE_SEED NodeType = iota
	NODE_TYPE_FEED
	NODE_TYPE_WEBSITE
	NODE_TYPE_BLOGROLL
)

type T_NOTHING = struct{}

var NOTHING = T_NOTHING{}
