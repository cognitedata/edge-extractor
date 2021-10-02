package iam

type SecurityCategoryListResponse struct {
	Items      SecurityCategoryList
	NextCursor string
}

type SecurityCategory struct {
	ID   uint64 `json:"id"`
	Name string `json:"name"`
}

type SecurityCategoryList []SecurityCategory

func (securityCategoryList *SecurityCategoryList) ConvertToIDList() SecurityCategoryIDList {
	var securityCategoryIDList SecurityCategoryIDList
	for _, s := range *securityCategoryList {
		securityCategoryIDList = append(securityCategoryIDList, s.ID)
	}
	return securityCategoryIDList
}

func (securityCategoryList *SecurityCategoryList) ConvertToCreateSecurityCategories() CreateSecurityCategorys {
	var createSecurityCategoryList CreateSecurityCategoryList
	for _, s := range *securityCategoryList {
		createSecurityCategory := CreateSecurityCategory{
			Name: s.Name,
		}
		createSecurityCategoryList = append(createSecurityCategoryList, createSecurityCategory)
	}
	return CreateSecurityCategorys{
		Items: createSecurityCategoryList,
	}
}

func (securityCategoryList *SecurityCategoryList) ConvertToItemsList() SecurityCategoryIDs {
	return SecurityCategoryIDs{
		Items: securityCategoryList.ConvertToIDList(),
	}
}

type CreateSecurityCategory struct {
	Name string `json:"name"`
}

type CreateSecurityCategoryList []CreateSecurityCategory

type CreateSecurityCategorys struct {
	Items CreateSecurityCategoryList `json:"items"`
}

type SecurityCategoryIDList []uint64

type SecurityCategoryIDs struct {
	Items SecurityCategoryIDList `json:"items"`
}
