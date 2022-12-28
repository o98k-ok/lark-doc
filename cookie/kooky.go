package cookie

import (
	"github.com/zellyn/kooky"
	_ "github.com/zellyn/kooky/browser/all" // register cookie store finders!
)

type Kooky struct {
	Domain     string
	CookieName string
}

func NewKooky(domain string, name string) Kooky {
	return Kooky{
		Domain:     domain,
		CookieName: name,
	}
}

func (k Kooky) Filter() []CookieItem {
	cookies := kooky.ReadCookies(kooky.Valid, kooky.DomainHasSuffix(k.Domain), kooky.Name(k.CookieName))

	var res []CookieItem
	for _, cookie := range cookies {
		ckie := CookieItem{
			Name:    cookie.Name,
			Value:   cookie.Value,
			Path:    cookie.Path,
			Domain:  cookie.Domain,
			Expires: cookie.Expires,
		}
		res = append(res, ckie)
	}
	return res
}
