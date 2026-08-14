package controllers

import "github.com/gin-gonic/gin"

// ExampleController 是一个最小化的示例控制器，仅用于演示 controller 的基本写法。
type ExampleController struct{}

func NewExampleController() *ExampleController {
	return &ExampleController{}
}

// Hello 返回一条简单的问候信息。
func (ec *ExampleController) Hello(c *gin.Context) {
	respondOK(c, "Example controller response", gin.H{
		"message": "Hello from ExampleController!",
	})
}
