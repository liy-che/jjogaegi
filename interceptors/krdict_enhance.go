package interceptors

import (
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/liy-che/jjogaegi/pkg"
	"gopkg.in/xmlpath.v2"
)

func KrDictEnhance(item *pkg.Item, options map[string]string) error {
	// format: strings.Join([]string{"krdict", lang, entryId, lexicalUnit}, ":")
	idSplit := strings.Split(item.ExternalID, ":")
	if len(idSplit) != 4 || idSplit[0] != "krdict" {
		return nil
	}
	entryID := idSplit[2]

	entry, err := fetchEntryNode(entryID, options)
	if err != nil {
		return err
	}

	switch len(item.Examples) {
	case 0:
		item.Examples = make([]pkg.Translation, 2, 2)
	case 1:
		item.Examples = append(item.Examples, pkg.Translation{})
	}

	// if the searched is a idiom, aka not a single word
	if strings.Contains(item.Hangul, " ") {
		query := fmt.Sprintf("//*[contains(subword, '%s')]/subsense_info", item.Hangul)
		if item.Def.Korean == "" {
			q := query + "/definition"
			item.Def.Korean = pkg.XpathString(entry, q)
		}
		if item.Def.English == "" {
			q := query + "/translation"
			item.Def.English = getEnglishDefinition(entry, q)
		}

		idx := 0
		q := query + "/example_info"
		if item.Examples[0].Korean == "" {
			item.Examples[0], idx = getExample(entry, "문장", q, idx)
		}
	
		if item.Examples[1].Korean == "" && idx != -1 {
			item.Examples[1], _ = getExample(entry, "문장", q, idx)
		}
		return nil
	}

	if item.Hangul == "" {
		item.Hangul = pkg.XpathString(entry, "/channel/item/word_info/word")
	}

	if item.Hanja == "" {
		item.Hanja = removeDuplicateCharacters(pkg.XpathString(entry, "/channel/item/word_info/original_language_info[language_type='한자']/original_language"))
	}

	if item.Pronunciation == "" {
		item.Pronunciation = pkg.XpathString(entry, "/channel/item/word_info/pronunciation_info/pronunciation")
	}

	// TODO: broken because missing
	// if item.AudioTag == "" || strings.HasPrefix(item.ItemTag, "[sound:say-"") {
	// }

	if item.Def.Korean == "" {
		item.Def.Korean = pkg.XpathString(entry, "/channel/item/word_info/sense_info/definition")
	}

	if item.Def.English == "" {
		item.Def.English = getEnglishDefinition(entry, "/channel/item/word_info/sense_info/translation")
	}

	if item.Antonym == "" {
		item.Antonym = pkg.XpathString(entry, "/channel/item/word_info/sense_info/rel_info[type='반대말']/word")
	}
	
	if item.Examples[0].Korean == "" {
		item.Examples[0], _ = getExample(entry, "구", "/channel/item/word_info/sense_info/example_info", 0)
	}

	if item.Examples[1].Korean == "" {
		item.Examples[1], _ = getExample(entry, "문장", "/channel/item/word_info/sense_info/example_info", 0)
	}

	if item.ImageTag == "" {
		// TODO: why isn't filter working?
		// get(entry, "/channel/item/word_info/sense_info/multimedia_info[type='사진']/link")
		item.ImageTag = pkg.XpathString(entry, "/channel/item/word_info/sense_info/multimedia_info/link")
		pkg.Debug(options, "at=enhance.image tag=%q", item.ImageTag)
	}

	if item.Grade == "" || item.Grade == "없음" {
		item.Grade = getWordGrade(entry)
	}

	return nil
}

func fetchEntryNode(entryID string, options map[string]string) (*xmlpath.Node, error) {
	url := fmt.Sprintf(
		"%s/api/view?key=%s&type_search=view&method=TARGET_CODE&part=word&q=%s&sort=dict&translated=y&trans_lang=1",
		options[pkg.OPT_KRDICT_API_URL],
		options[pkg.OPT_KRDICT_API_KEY],
		entryID,
	)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		log.Printf("download type=eng url=%q err=%q", url, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Non-200 viewing KR DICT API")
	}

	return xmlpath.Parse(resp.Body)
}

func getEnglishDefinition(node *xmlpath.Node, transPath string) string {
	transWord := pkg.XpathString(node, transPath+"/trans_word")
	transDfn := pkg.XpathString(node, transPath+"/trans_dfn")

	return transWord + " := " + transDfn
}

func getWordGrade(node *xmlpath.Node) string {
	grade := pkg.XpathString(node, "/channel/item/word_info/word_grade")
	switch grade {
	case "없음":
		return ""
	default:
		// TODO: consider changing to numbers
		return grade
	}
}

func getExample(node *xmlpath.Node, exampleType string, examplePath string, loc int) (pkg.Translation, int) {
	examplesIter := xmlpath.MustCompile(examplePath).Iter(node)
	idx := 1
	for {
		if !examplesIter.Next() {
			break
		}

		exampleNode := examplesIter.Node()

		if pkg.XpathString(exampleNode, "type") == exampleType && idx > loc {
			return pkg.Translation{Korean: pkg.XpathString(exampleNode, "example")}, idx
		}

		idx += 1
	}
	return pkg.Translation{}, -1
}

func removeDuplicateCharacters(text string) string {
	var lastRune rune
	var parts []rune
	runes := []rune(text)

	for i := 0; i < len(runes); i++ {
		c := runes[i]
		if lastRune != c {
			parts = append(parts, c)
		}

		lastRune = c
	}

	return string(parts)
}