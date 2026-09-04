package utils

import (
	"fmt"
	"strconv"
)

func ParsePagination(pageValue, pageSizeValue string) (page, pageSize int, err error) {
	page, pageSize = 1, 20
	if pageValue != "" {
		page, err = strconv.Atoi(pageValue)
		if err != nil || page < 1 {
			return 0, 0, fmt.Errorf("page must be a positive integer")
		}
	}
	if pageSizeValue != "" {
		pageSize, err = strconv.Atoi(pageSizeValue)
		if err != nil || pageSize < 1 || pageSize > 100 {
			return 0, 0, fmt.Errorf("page_size must be between 1 and 100")
		}
	}
	return page, pageSize, nil
}
