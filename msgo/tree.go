package msgo

import "strings"

// treeNode 结构体表示路由树的节点
type treeNode struct {
	name       string      // 节点名称
	children   []*treeNode // 子节点列表
	routerName string      // 路由名称，记录从根节点到当前节点的路径
	isEnd      bool        // 是否为叶子节点
}

// Put 方法用于向路由树中添加路由路径
func (t *treeNode) Put(path string) {
	root := t
	strs := strings.Split(path, "/")
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == name {
				isMatch = true
				t = node
				break
			}
		}
		if !isMatch {
			isEnd := false
			if index == len(strs)-1 {
				isEnd = true
			}
			node := &treeNode{name: name, children: make([]*treeNode, 0), isEnd: isEnd}
			children = append(children, node)
			t.children = children
			t = node
		}
	}
	t = root
}

// Get 方法用于根据给定的路径查找对应的路由节点
func (t *treeNode) Get(path string) *treeNode {
	strs := strings.Split(path, "/")
	routerName := ""
	for index, name := range strs {
		if index == 0 {
			continue
		}
		children := t.children
		isMatch := false
		for _, node := range children {
			if node.name == name ||
				node.name == "*" ||
				strings.Contains(node.name, ":") {
				isMatch = true
				routerName += "/" + node.name
				node.routerName = routerName
				t = node
				if index == len(strs)-1 {
					return node
				}
				break
			}
		}
		if !isMatch {
			for _, node := range children {
				// /user/**
				// /user/get/userInfo
				// /user/aa/bb
				if node.name == "**" {
					routerName += "/" + node.name
					node.routerName = routerName
					return node
				}
			}

		}
	}
	return nil
}
