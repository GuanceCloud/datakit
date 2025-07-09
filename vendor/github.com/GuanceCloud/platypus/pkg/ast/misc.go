package ast

type CommaExpr struct {
	List []*Node
}

func (e *CommaExpr) Append(n *Node) {
	e.List = append(e.List, n)
}

type FnCallParams struct {
	List []*Node
}

func (c *FnCallParams) Append(n *Node) {
	c.List = append(c.List, n)
}
