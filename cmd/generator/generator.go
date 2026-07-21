package main

import (
	"bbs-go/internal/models"

	"github.com/mlogclub/codegen"
)

func main() {
	codegen.Generate(
		"./",
		"bbs-go",
		1,
		codegen.GetGenerateStruct(&models.Migration{}),
		codegen.GetGenerateStruct(&models.Vote{}),
		codegen.GetGenerateStruct(&models.VoteOption{}),
		codegen.GetGenerateStruct(&models.VoteRecord{}),
	)
}
