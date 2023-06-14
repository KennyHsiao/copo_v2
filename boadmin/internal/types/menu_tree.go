package types

import (
	"encoding/json"
)

type RoleQueryResponseX struct {
	Role
	MenuTree json.RawMessage `json:"menuTree"`
}

func GenMenuTree(menus []Menu) []*MenuTree {
	var tmp = make([]*MenuTree, 0)
	toTree(menus, &tmp)
	return tmp
}

func toTree(menus []Menu, tmp *[]*MenuTree) {
	for _, m := range menus {
		if m.ParentID == 0 {
			child := make([]*MenuTree, 0)
			*tmp = append(*tmp, &MenuTree{
				Menu:  m,
				Child: child,
			})
		} else {
			for _, mm := range *tmp {
				if mm == nil {
					continue
				}
				if m.ParentID == mm.ID {
					mm.Child = append(mm.Child, &MenuTree{
						Menu: m,
					})
				} else {
					toTree([]Menu{m}, &mm.Child)
				}
			}
		}
	}
}

func GenMenuTreeFilter(menus []Menu, permitMap map[int64]bool) []*MenuTree {
	var tmp = make([]*MenuTree, 0)
	toTreeFilter(menus, &tmp, permitMap)
	return tmp
}

func toTreeFilter(menus []Menu, tmp *[]*MenuTree, permitMap map[int64]bool) {
	for _, m := range menus {
		if m.ParentID == 0 {
			child := make([]*MenuTree, 0)

			m.Permits = filterPermit(m.Permits, permitMap)

			*tmp = append(*tmp, &MenuTree{
				Menu:  m,
				Child: child,
			})
		} else {
			for _, mm := range *tmp {
				if mm == nil {
					continue
				}
				if m.ParentID == mm.ID {

					m.Permits = filterPermit(m.Permits, permitMap)

					mm.Child = append(mm.Child, &MenuTree{
						Menu: m,
					})
				} else {
					toTreeFilter([]Menu{m}, &mm.Child, permitMap)
				}
			}
		}
	}
}

// filterPermit 過濾userPermits
func filterPermit(permits []Permit, permitMap map[int64]bool) []Permit {
	filterPermits := []Permit{}
	for _, ps := range permits {
		if _, ok := permitMap[ps.ID]; ok {
			filterPermits = append(filterPermits, ps)
		}
	}

	return filterPermits
}
