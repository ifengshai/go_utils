package main

import (
	"log"
	"net/http"

	"gin_test/go_utils"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)
func main() {

	//这部分可放在main.go或router.go中
	//初始化翻译器
	if err := go_utils.InitTrans("zh"); err != nil {
		log.Fatalf("init trans failed, err:%v\n", err)
		return
	}
	//将我们自定义的校验方法注册到 validator中
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("notadmin", NotAdmin)
	}

	type TestApiApi struct {
		Age int `form:"age" binding:"required"`
		//在参数 binding 上使用自定义的校验方法函数注册时候的名称
		Name string `form:"name" binding:"required,notadmin"`
	}


	r := gin.Default()
	r.GET("/bookable", func(c *gin.Context) {

		req := TestApiApi{}
		err := c.ShouldBindWith(&req, binding.Query)//绑定query
		if err != nil {
			c.JSON(http.StatusBadRequest, go_utils.ErrResp(err)) //这里调用翻译
			return
		}
		c.String(http.StatusOK, "通过参数验证了")
	})
	//监听端口默认为8080
	r.Run(":8000")
}

//自定义的校验方法
func NotAdmin(fl validator.FieldLevel) bool {
	if fl.Field().String() == "admin" {
		return false
	}
	return true
}