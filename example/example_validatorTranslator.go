package main

import (
	ut "github.com/go-playground/universal-translator"
	"github.com/ifengshai/go_utils"
	"log"
	"net/http"

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
		v.RegisterValidation("notadmina", NotAdmin)
	}

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 添加额外自定义验证方法报错的翻译
		_ = v.RegisterTranslation("notadmina", go_utils.Trans, func(ut ut.Translator) error {
			return ut.Add("notadmina", "{0} adminaaaa不能用!", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			t, _ := ut.T("notadmina", fe.Field())
			return t
		})
	}

	type TestApiApi struct {
		Age int `form:"age" binding:"required"`
		//在参数 binding 上使用自定义的校验方法函数注册时候的名称
		Name string `form:"name" binding:"required,notadmina"`
	}

	r := gin.Default()
	r.GET("/bookable", func(c *gin.Context) {

		req := TestApiApi{}
		err := c.ShouldBindWith(&req, binding.Query) //绑定query
		if err != nil {
			c.JSON(http.StatusBadRequest, go_utils.ErrResp(err)) //这里调用翻译
			return
		}
		c.String(http.StatusOK, "通过参数验证了")
	})
	//监听端口默认为8080
	r.Run(":8000")
}

// NotAdmin 自定义的校验方法
func NotAdmin(fl validator.FieldLevel) bool {
	if fl.Field().String() == "admin" {
		return false
	}
	return true
}
