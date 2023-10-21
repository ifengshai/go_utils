package main

import (
	"encoding/json"
	"fmt"
)

type Node struct {
	Id       int     `json:"id"`
	ParentId int     `json:"parent_id"`
	Name     string  `json:"name"`
	Children []*Node `json:"children"`
}

func getTreeIterative(list []*Node, parentId int) []*Node {
	memo := make(map[int]*Node)
	for _, v := range list {
		if _, ok := memo[v.Id]; ok {
			v.Children = memo[v.Id].Children
			memo[v.Id] = v
		} else {
			v.Children = make([]*Node, 0)
			memo[v.Id] = v
		}
		if _, ok := memo[v.ParentId]; ok {
			memo[v.ParentId].Children = append(memo[v.ParentId].Children, memo[v.Id])
		} else {
			memo[v.ParentId] = &Node{Children: []*Node{memo[v.Id]}}
		}
	}
	return memo[parentId].Children
}

func main() {
	list := []*Node{
		{4, 3, "ABA", nil},
		{3, 1, "AB", nil},
		{1, 0, "A", nil},
		{2, 1, "AA", nil},
		{5, 3, "ABB", nil},
	}
	res := getTreeIterative(list, 0)
	bytes, _ := json.MarshalIndent(res, "", "    ")
	fmt.Printf("%s\n", bytes)
}
