package service

import (
	"OJ/internal/define"
	"OJ/internal/helper"
	"OJ/internal/models"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetProblemList
// @Tags 公共方法
// @Summary 问题列表
// @Param page query int false "page"
// @Param size query int false "size"
// @Param keyword query string false "keyword"
// @Param category_identity query string false "category_identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-list [get]
func GetProblemList(c *gin.Context) {
	size, _ := strconv.Atoi(c.DefaultQuery("size", define.DefaultSize))
	page, err := strconv.Atoi(c.DefaultQuery("page", define.DefaultPage))
	if err != nil {
		log.Println("GetProblemList Page strconv Error:", err)
		return
	}
	page = (page - 1) * size
	var count int64
	keyword := c.Query("keyword")
	categoryIdentity := c.Query("category_identity")

	list := make([]*models.ProblemBasic, 0)
	// 在接口函数中调用操作数据模型的函数
	err = models.GetProblemList(keyword, categoryIdentity).Distinct("`problem_basic`.`id`").Count(&count).Error
	if err != nil {
		log.Println("GetProblemList Count Error:", err)
		return
	}
	// 在接口函数中调用操作数据模型的函数
	err = models.GetProblemList(keyword, categoryIdentity).Offset(page).Limit(size).Find(&list).Error
	if err != nil {
		log.Println("Get Problem List Error:", err)
		return
	}
	// 将后端查询到的数据返回给前端
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"list":  list,
			"count": count,
		},
	})
}

// GetProblemDetail
// @Tags 公共方法
// @Summary 问题详情
// @Param identity query string false "problem identity"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /problem-detail [get]
func GetProblemDetail(c *gin.Context) {
	identity := c.Query("identity")
	if identity == "" {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "问题唯一标识不能为空",
		})
		return
	}
	// 等价于代码 data := models.ProblemBasic{} 初始化一个空的结构体实例
	// 值传递类型 new  引用传递类型 make
	data := new(models.ProblemBasic)
	err := models.DB.Where("identity = ?", identity).
		Preload("ProblemCategories").Preload("ProblemCategories.CategoryBasic").
		First(&data).Error
	if err != nil {
		// ErrRecordNotFound 记录未找到错误
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, gin.H{
				"code": -1,
				"msg":  "问题不存在",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Get ProblemDetail Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": data,
	})
}

// ProblemCreate
// @Tags 管理员私有方法
// @Summary 问题创建
// @Accept json
// @Param authorization header string true "authorization"
// @Param data body define.ProblemBasic true "ProblemBasic"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-create [post]
func ProblemCreate(c *gin.Context) {
	in := new(define.ProblemBasic)
	err := c.ShouldBindJSON(in)
	if err != nil {
		log.Println("[JsonBind Error] : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数错误",
		})
		return
	}
	// 前端post请求携带过来的数据  由c.ShouldBindJSON(in) 将传过来的数据绑定在ProblemBasic的实例对象上
	if in.Title == "" || in.Content == "" || len(in.ProblemCategories) == 0 || len(in.TestCases) == 0 || in.MaxRuntime == 0 || in.MaxMem == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}
	identity := helper.GetUUID()

	// 初始化models.ProblemBasis的实例对象
	data := &models.ProblemBasic{
		Identity:   identity,
		Title:      in.Title,
		Content:    in.Content,
		MaxRuntime: in.MaxRuntime,
		MaxMem:     in.MaxMem,
		CreatedAt:  models.MyTime(time.Now()),
		UpdatedAt:  models.MyTime(time.Now()),
	}
	// 处理分类
	categoryBasics := make([]*models.ProblemCategory, 0)
	for _, id := range in.ProblemCategories {
		categoryBasics = append(categoryBasics, &models.ProblemCategory{
			ProblemId:  data.ID,
			CategoryId: uint(id),
			CreatedAt:  models.MyTime(time.Now()),
			UpdatedAt:  models.MyTime(time.Now()),
		})
	}
	data.ProblemCategories = categoryBasics
	// 处理测试用例
	testCaseBasics := make([]*models.TestCase, 0)
	for _, v := range in.TestCases {
		// 举个例子 {"input":"1 2\n","output":"3\n"}
		testCaseBasic := &models.TestCase{
			Identity:        helper.GetUUID(),
			ProblemIdentity: identity,
			Input:           v.Input,
			Output:          v.Output,
			CreatedAt:       models.MyTime(time.Now()),
			UpdatedAt:       models.MyTime(time.Now()),
		}
		testCaseBasics = append(testCaseBasics, testCaseBasic)
	}
	data.TestCases = testCaseBasics

	// 创建问题
	err = models.DB.Create(data).Error
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Create Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"data": map[string]interface{}{
			"identity": data.Identity,
		},
	})
}

// ProblemModify
// @Tags 管理员私有方法
// @Summary 问题修改
// @Param authorization header string true "authorization"
// @Param data body define.ProblemBasic true "ProblemBasic"
// @Success 200 {string} json "{"code":"200","data":""}"
// @Router /admin/problem-modify [put]
func ProblemModify(c *gin.Context) {
	in := new(define.ProblemBasic)
	err := c.ShouldBindJSON(in)
	if err != nil {
		log.Println("[JsonBind Error] : ", err)
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数错误",
		})
		return
	}
	if in.Identity == "" || in.Title == "" || in.Content == "" || len(in.ProblemCategories) == 0 || len(in.TestCases) == 0 || in.MaxRuntime == 0 || in.MaxMem == 0 {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "参数不能为空",
		})
		return
	}

	if err := models.DB.Transaction(func(tx *gorm.DB) error {
		// 问题基础信息保存 problem_basic
		problemBasic := &models.ProblemBasic{
			Identity:   in.Identity,
			Title:      in.Title,
			Content:    in.Content,
			MaxRuntime: in.MaxRuntime,
			MaxMem:     in.MaxMem,
			UpdatedAt:  models.MyTime(time.Now()),
		}
		err := tx.Where("identity = ?", in.Identity).Updates(problemBasic).Error
		if err != nil {
			return err
		}
		// 查询问题详情
		err = tx.Where("identity = ?", in.Identity).Find(problemBasic).Error
		if err != nil {
			return err
		}

		// 关联问题分类的更新
		// 1、删除已存在的关联关系
		err = tx.Where("problem_id = ?", problemBasic.ID).Delete(new(models.ProblemCategory)).Error
		if err != nil {
			return err
		}
		// 2、新增新的关联关系
		pcs := make([]*models.ProblemCategory, 0)
		for _, id := range in.ProblemCategories {
			pcs = append(pcs, &models.ProblemCategory{
				ProblemId:  problemBasic.ID,
				CategoryId: uint(id),
				CreatedAt:  models.MyTime(time.Now()),
				UpdatedAt:  models.MyTime(time.Now()),
			})
		}
		err = tx.Create(&pcs).Error
		if err != nil {
			return err
		}
		// 关联测试案例的更新
		// 1、删除已存在的关联关系
		err = tx.Where("problem_identity = ?", in.Identity).Delete(new(models.TestCase)).Error
		if err != nil {
			return err
		}
		// 2、增加新的关联关系
		tcs := make([]*models.TestCase, 0)
		for _, v := range in.TestCases {
			// 举个例子 {"input":"1 2\n","output":"3\n"}
			tcs = append(tcs, &models.TestCase{
				Identity:        helper.GetUUID(),
				ProblemIdentity: in.Identity,
				Input:           v.Input,
				Output:          v.Output,
				CreatedAt:       models.MyTime(time.Now()),
				UpdatedAt:       models.MyTime(time.Now()),
			})
		}
		err = tx.Create(tcs).Error
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code": -1,
			"msg":  "Problem Modify Error:" + err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"code": 200,
		"msg":  "问题修改成功",
	})
}
