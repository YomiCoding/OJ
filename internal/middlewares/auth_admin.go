package middlewares

import (
	"gin-gorm-oj-master/helper"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthAdminCheck is a middleware function that checks if the user is authenticated with admin role.
func AuthAdminCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		userClaim, err := helper.AnalyseToken(auth)
		if err != nil {
			// 不会停止当前处理程序  调用 Abort 以确保不调用此请求的其它中间件和后续处理程序
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Authorization",
			})
			return
		}
		if userClaim == nil || userClaim.IsAdmin != 1 {
			c.Abort()
			c.JSON(http.StatusOK, gin.H{
				"code": http.StatusUnauthorized,
				"msg":  "Unauthorized Admin",
			})
			return
		}
		// 洋葱模型  c.Next()后面的代码暂时不执行，先去执行接下来的中间件和该请求的后续处理程序，待返回数据到前端时在执行c.Next()后面未执行的代码
		c.Next()
	}
}
