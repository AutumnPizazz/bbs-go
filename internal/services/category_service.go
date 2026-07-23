package services

import (
	"errors"
	"strings"

	"bbs-go/internal/models"
	"bbs-go/internal/models/constants"
	"bbs-go/internal/models/dto"
	"bbs-go/internal/pkg/locales"

	"bbs-go/internal/pkg/params"

	"github.com/mlogclub/simple/sqls"

	"bbs-go/internal/repositories"

	"gorm.io/gorm"
)

var CategoryService = newCategoryService()

func newCategoryService() *categoryService {
	return &categoryService{}
}

type categoryService struct {
}

func (s *categoryService) Get(id int64) *models.Category {
	return repositories.CategoryRepository.Get(sqls.DB(), id)
}

func (s *categoryService) Take(where ...interface{}) *models.Category {
	return repositories.CategoryRepository.Take(sqls.DB(), where...)
}

func (s *categoryService) Find(cnd *sqls.Cnd) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), cnd)
}

func (s *categoryService) FindOne(cnd *sqls.Cnd) *models.Category {
	return repositories.CategoryRepository.FindOne(sqls.DB(), cnd)
}

func (s *categoryService) FindPageByParams(params *params.QueryParams) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByParams(sqls.DB(), params)
}

func (s *categoryService) FindPageByCnd(cnd *sqls.Cnd) (list []models.Category, paging *sqls.Paging) {
	return repositories.CategoryRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *categoryService) Create(t *models.Category) error {
	err := repositories.CategoryRepository.Create(sqls.DB(), t)
	if err == nil {
		ContentAccessService.InvalidateAll()
	}
	return err
}

func (s *categoryService) Update(t *models.Category) error {
	err := repositories.CategoryRepository.Update(sqls.DB(), t)
	if err == nil {
		ContentAccessService.InvalidateAll()
	}
	return err
}

func (s *categoryService) Updates(id int64, columns map[string]interface{}) error {
	err := repositories.CategoryRepository.Updates(sqls.DB(), id, columns)
	if err == nil {
		ContentAccessService.InvalidateAll()
	}
	return err
}

func (s *categoryService) UpdateColumn(id int64, name string, value interface{}) error {
	err := repositories.CategoryRepository.UpdateColumn(sqls.DB(), id, name, value)
	if err == nil {
		ContentAccessService.InvalidateAll()
	}
	return err
}

// GetAttachmentConfig returns the effective attachment policy for a category.
// A configured child overrides its ancestors; otherwise the global default is used.
func (s *categoryService) GetAttachmentConfig(categoryId int64) dto.AttachmentConfig {
	cfg := SysConfigService.GetAttachmentConfig()
	for categoryId > 0 {
		category := s.Get(categoryId)
		if category == nil {
			break
		}
		if category.AttachmentPolicyConfigured {
			cfg.Enabled = category.AttachmentEnabled
			cfg.MaxSizeMB = category.AttachmentMaxSizeMB
			cfg.MaxCount = category.AttachmentMaxCount
			if types := parseAttachmentAllowedTypes(category.AttachmentAllowedTypes); len(types) > 0 {
				cfg.AllowedTypes = types
			}
			return cfg
		}
		categoryId = category.ParentId
	}
	return cfg
}

func (s *categoryService) ValidateAttachmentPolicy(category *models.Category) error {
	if category == nil {
		return errors.New("category is required")
	}
	if category.AttachmentMaxSizeMB < 0 {
		return errors.New("attachment max size must be greater than or equal to 0")
	}
	if category.AttachmentMaxCount < 0 {
		return errors.New("attachment max count must be greater than or equal to 0")
	}
	return nil
}

func parseAttachmentAllowedTypes(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n' || r == '\r' || r == '\t' || r == ' '
	})
	ret := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			continue
		}
		if part != "*" && part != "*/*" && !strings.HasPrefix(part, ".") {
			part = "." + part
		}
		if _, ok := seen[part]; ok {
			continue
		}
		seen[part] = struct{}{}
		ret = append(ret, part)
	}
	return ret
}

// DeleteWithCheck 删除节点，若为一级且有子节点则返回错误
func (s *categoryService) DeleteWithCheck(id int64) error {
	category := s.Get(id)
	if category == nil {
		return nil
	}
	if category.ParentId == 0 {
		children := s.GetChildren(id)
		if len(children) > 0 {
			return errors.New(locales.Get("topic.category.has_children"))
		}
	}
	err := repositories.CategoryRepository.Updates(sqls.DB(), id, map[string]interface{}{
		"status": constants.StatusDeleted,
	})
	if err == nil {
		ContentAccessService.InvalidateAll()
	}
	return err
}

// GetTopLevelCategories 仅一级节点（parent_id=0），用于导航
func (s *categoryService) GetTopLevelCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", 0).
		Asc("sort_no").Desc("id"))
}

// GetChildren 获取某一级下的二级节点
func (s *categoryService) GetChildren(parentId int64) []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().
		Eq("status", constants.StatusOk).
		Eq("parent_id", parentId).
		Asc("sort_no").Desc("id"))
}

// GetCategoryIdsForList returns a category and all active descendants.
func (s *categoryService) GetCategoryIdsForList(categoryId int64) []int64 {
	if categoryId <= 0 {
		return nil
	}
	categories := s.GetCategories()
	byParent := make(map[int64][]int64)
	found := false
	for _, category := range categories {
		if category.Id == categoryId {
			found = true
		}
		byParent[category.ParentId] = append(byParent[category.ParentId], category.Id)
	}
	if !found {
		return nil
	}
	ids := []int64{categoryId}
	queue := []int64{categoryId}
	for len(queue) > 0 {
		parentId := queue[0]
		queue = queue[1:]
		for _, childId := range byParent[parentId] {
			ids = append(ids, childId)
			queue = append(queue, childId)
		}
	}
	return ids
}

func (s *categoryService) GetCategories() []models.Category {
	return repositories.CategoryRepository.Find(sqls.DB(), sqls.NewCnd().Eq("status", constants.StatusOk).Asc("sort_no").Desc("id"))
}

func (s *categoryService) GetNextSortNo() int {
	if max := s.FindOne(sqls.NewCnd().Eq("status", constants.StatusOk).Desc("sort_no")); max != nil {
		return max.SortNo + 1
	}
	return 0
}

func (s *categoryService) UpdateSort(ids []int64) error {
	return sqls.DB().Transaction(func(tx *gorm.DB) error {
		for i, id := range ids {
			if err := repositories.CategoryRepository.UpdateColumn(tx, id, "sort_no", i); err != nil {
				return err
			}
		}
		return nil
	})
}
