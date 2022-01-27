package form

import (
	"fmt"
	"sync"
)

type DataToFill struct {
	Pages []Page

	m    sync.Mutex
	form map[string]interface{}
}

type Page struct {
	PageNumber int
	Name       string
	ID         string
	Box        int
	Rows       [][]interface{}
}

func (d *DataToFill) CreateForm() (map[string]interface{}, error) {
	var wg sync.WaitGroup
	d.form = make(map[string]interface{})

	if len(d.Pages) > 2 {
		return nil, fmt.Errorf("pages cannot be more than 2")
	}

	for i := range d.Pages {
		wg.Add(1)
		var err error
		go func(i int, form map[string]interface{}, m *sync.Mutex) {
			defer wg.Done()
			err = d.Pages[i].createPage(form, m)
			if err != nil {
				return
			}
		}(i, d.form, &d.m)
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()
	return d.form, nil
}

func (p *Page) createPage(form map[string]interface{}, m *sync.Mutex) error {
	if p.Box < 1 || p.Box > 3 {
		return fmt.Errorf("the box parameter must take values from 1 to 3")
	}
	if len(p.Rows) > 14 {
		return fmt.Errorf("rows cannot be more than 14")
	}

	dTotal, eTotal, gTotal, hTotal, err := p.createRows(form, m)
	if err != nil {
		return err
	}

	m.Lock()
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_1[0]", p.PageNumber, p.PageNumber)] = p.Name
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_2[0]", p.PageNumber, p.PageNumber)] = p.ID
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].c%d_1[%d]", p.PageNumber, p.PageNumber, p.Box-1)] = p.Box

	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_115[0]", p.PageNumber, p.PageNumber)] = dTotal
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_116[0]", p.PageNumber, p.PageNumber)] = eTotal
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_118[0]", p.PageNumber, p.PageNumber)] = gTotal
	form[fmt.Sprintf("topmostSubform[0].Page%d[0].f%d_119[0]", p.PageNumber, p.PageNumber)] = hTotal
	m.Unlock()

	return nil
}

func (p *Page) createRows(form map[string]interface{}, m *sync.Mutex) (dTotal, eTotal, gTotal, hTotal float64, err error) {
	id := 3
	for i := range p.Rows {
		if len(p.Rows[i]) > 7 {
			return 0, 0, 0, 0, fmt.Errorf("columns cannot be more than 7")
		}
		var gain float64
		for j := 0; j < 7; j++ {
			if len(p.Rows[i]) > j {
				m.Lock()
				form[fmt.Sprintf("topmostSubform[0].Page%d[0].Table_Line1[0].Row%d[0].f%d_%d[0]", p.PageNumber, i+1, p.PageNumber, id)] = p.Rows[i][j]
				m.Unlock()

				if (j > 2 && j < 5) || j == 6 {
					val, ok := p.Rows[i][j].(float64)
					if !ok {
						return 0, 0, 0, 0, fmt.Errorf("costs should be float64")
					}
					switch j {
					case 3:
						gain = val
						dTotal += val
					case 4:
						gain -= val
						eTotal += val
					case 6:
						gain += val
						gTotal += val
					}
				}
			}
			id++
		}
		m.Lock()
		form[fmt.Sprintf("topmostSubform[0].Page%d[0].Table_Line1[0].Row%d[0].f%d_%d[0]", p.PageNumber, i+1, p.PageNumber, id)] = gain
		m.Unlock()

		hTotal += gain
		id++
	}
	return dTotal, eTotal, gTotal, hTotal, nil
}
