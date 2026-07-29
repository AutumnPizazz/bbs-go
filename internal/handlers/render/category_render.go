package render

import (
	"bbs-go/internal/models"
	"bbs-go/internal/models/resp"
	"bbs-go/internal/services"

	"github.com/mlogclub/simple/common/strs"
)

func BuildCategory(category *models.Category) *resp.CategoryResponse {
	if category == nil {
		return nil
	}
	if strs.IsBlank(category.Logo) {
		category.Logo = "/res/images/category_default.svg"
	}
	return &resp.CategoryResponse{
		Id:               category.Id,
		ParentId:         category.ParentId,
		Name:             category.Name,
		Logo:             category.Logo,
		Description:      category.Description,
		AttachmentConfig: services.CategoryService.GetAttachmentConfig(category.Id),
	}
}

func BuildCategoryWithChildren(category *models.Category, allowedIds []int64) *resp.CategoryResponse {
	r := BuildCategory(category)
	if r == nil {
		return nil
	}
	// Always populate direct children for any level (not just root).
	// This enables third-level (and deeper) category navigation in the frontend.
	children := services.CategoryService.GetChildren(category.Id)
	if len(children) > 0 {
		r.Children = BuildCategoryResponses(children)
		// Populate stats for child categories so the frontend can show count badges.
		for i := range r.Children {
			stats := services.TopicService.GetCategoryStats(r.Children[i].Id, allowedIds)
			r.Children[i].TopicCount = stats.TopicCount
			r.Children[i].QaCount = stats.QaCount
			r.Children[i].SolvedCount = stats.SolvedCount
			r.Children[i].UnsolvedCount = stats.UnsolvedCount
		}
	}
	return r
}

func BuildCategoryResponses(categories []models.Category) []resp.CategoryResponse {
	if len(categories) == 0 {
		return nil
	}
	var ret []resp.CategoryResponse
	for _, category := range categories {
		ret = append(ret, *BuildCategory(&category))
	}
	return ret
}

func BuildCategoryResponseTree(parentId int64, list []models.Category) []resp.CategoryResponse {
	var ret []resp.CategoryResponse
	for _, category := range list {
		if category.ParentId != parentId {
			continue
		}
		item := BuildCategory(&category)
		if item == nil {
			continue
		}
		children := BuildCategoryResponseTree(category.Id, list)
		if len(children) > 0 {
			item.Children = children
		}
		ret = append(ret, *item)
	}
	return ret
}

// PopulateCategoryStats fills topic/qa/solved/unsolved counts for a category tree.
func PopulateCategoryStats(tree []resp.CategoryResponse, allowedIds []int64) {
	for i := range tree {
		stats := services.TopicService.GetCategoryStats(tree[i].Id, allowedIds)
		tree[i].TopicCount = stats.TopicCount
		tree[i].QaCount = stats.QaCount
		tree[i].SolvedCount = stats.SolvedCount
		tree[i].UnsolvedCount = stats.UnsolvedCount
		if len(tree[i].Children) > 0 {
			PopulateCategoryStats(tree[i].Children, allowedIds)
		}
	}
}

// AggregateCategoryStats bottom-up aggregates children's stats into each parent,
// so parent nodes reflect totals including all descendants (recursive).
// Call after PopulateCategoryStats has filled per-node stats.
func AggregateCategoryStats(tree []resp.CategoryResponse) {
	for i := range tree {
		if len(tree[i].Children) > 0 {
			AggregateCategoryStats(tree[i].Children)
			for _, child := range tree[i].Children {
				tree[i].TopicCount += child.TopicCount
				tree[i].QaCount += child.QaCount
				tree[i].SolvedCount += child.SolvedCount
				tree[i].UnsolvedCount += child.UnsolvedCount
			}
		}
	}
}

func BuildCategoryTree(parentId int64, list []models.Category) []resp.CategoryTreeItem {
	var ret []resp.CategoryTreeItem
	for _, category := range list {
		if category.ParentId == parentId {
			children := BuildCategoryTree(category.Id, list)
			logo := category.Logo
			if strs.IsBlank(logo) {
				logo = "/res/images/category_default.svg"
			}
			ret = append(ret, resp.CategoryTreeItem{
				Id:          category.Id,
				ParentId:    category.ParentId,
				Name:        category.Name,
				Logo:        logo,
				Description: category.Description,
				SortNo:      category.SortNo,
				Status:      category.Status,
				CreateTime:  category.CreateTime,
				Children:    children,
			})
		}
	}
	return ret
}
