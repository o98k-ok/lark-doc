package doc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/o98k-ok/aggregation/cookie"
	np "github.com/o98k-ok/pcurl/http"
)

type Lark struct {
	Offset  int
	Count   int
	URL     string
	Client  *http.Client
	Session *string
}

func NewLark() *Lark {
	return &Lark{
		Offset: 0,
		Count:  10,
		URL:    "https://xndn97f8ch.feishu.cn/space/api/search/refine_search",
		Client: &http.Client{},
	}
}

func (l *Lark) WithPage(offset, count int) *Lark {
	l.Offset = offset
	l.Count = count
	return l
}

func (l *Lark) CustomSession(session string) *Lark {
	l.Session = &session
	return l
}

type Artitle struct {
	Title     string `json:"title"`
	Preview   string `json:"preview"`
	OpenTime  uint32 `json:"open_time"`
	EditTime  uint32 `json:"edit_time"`
	Author    string `json:"author"`
	URL       string `json:"url"`
	WikiInfos []struct {
		WikiUrl string `json:"wiki_url"`
	} `json:"wiki_infos"`
}

type ArtitleResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Entities struct {
			Artitles map[string]Artitle `json:"objs"`
		} `json:"entities"`
		HasMore bool `json:"has_more"`
		Total   int  `json:"total"`
	} `json:"data"`
}

func (l Lark) Query(query string) ([]Entity, error) {
	offset, count := strconv.FormatInt(int64(l.Offset), 10), strconv.FormatInt(int64(l.Count), 10)
	var session cookie.CookieItem
	if l.Session == nil || len(*l.Session) == 0 {
		sessions := cookie.NewKooky("feishu.cn", "session").Filter()
		if len(sessions) == 0 {
			return nil, errors.New("no fuch session")
		}
		session = sessions[0]
	} else {
		session.Name = "session"
		session.Value = *l.Session
	}

	resp, err := np.NewRequest(l.Client, l.URL).
		AddParam("query", query).AddParam("offset", offset).AddParam("count", count).
		AddHeader("cookie", fmt.Sprintf("%s=%s;", session.Name, session.Value)).Do()
	if err != nil {
		return nil, err
	}
	if resp.Code != http.StatusOK {
		return nil, fmt.Errorf("statu code %v", resp.Code)
	}

	defer resp.Body.Close()
	var artitle ArtitleResp
	if err = json.NewDecoder(resp.Body).Decode(&artitle); err != nil {
		return nil, err
	}

	if artitle.Code != 0 {
		return nil, fmt.Errorf("feishu resp code is %v", artitle.Code)
	}

	var res []Entity
	for _, artk := range artitle.Data.Entities.Artitles {
		var url string
		if len(artk.URL) != 0 {
			url = artk.URL
		}

		if len(artk.WikiInfos) != 0 {
			url = artk.WikiInfos[0].WikiUrl
		}
		entity := Entity{
			Title:  artk.Title,
			Intro:  artk.Preview,
			Url:    url,
			Owner:  artk.Author,
			ViewTS: artk.OpenTime,
			EditTs: artk.OpenTime,
		}
		res = append(res, entity)
	}
	return res, nil
}
