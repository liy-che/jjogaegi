package interceptors

import (
	"bufio"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/liy-che/jjogaegi/pkg"
	"gopkg.in/xmlpath.v2"
)

func NewKrDictLookup(interactiveIn io.Reader, interactiveOut io.Writer) pkg.InterceptorFunc {
	return func(item *pkg.Item, options map[string]string) error {
		return krDictLookupDoubleSpaced(interactiveIn, interactiveOut, item, options)
	}
}

func krDictLookupDoubleSpaced(in io.Reader, out io.Writer, item *pkg.Item, options map[string]string) error {
	err := krDictLookup(in, out, item, options)
	fmt.Fprintf(out, "\n\n")
	return err
}

func krDictLookup(in io.Reader, out io.Writer, item *pkg.Item, options map[string]string) error {
	if options[pkg.OPT_LOOKUP] != strconv.FormatBool(true) {
		return nil
	}

	interactive := options[pkg.OPT_INTERACTIVE] == strconv.FormatBool(true)

	if item.ExternalID != "" {
		return nil
	}

	q := item.Hangul
	if item.Hanja != "" {
		q = item.Hanja
	}

	results, err := search(q, options)
	if err != nil {
		return err
	}

	resultsIntr := xmlpath.MustCompile("/channel/item").Iter(results)
	choices := []*xmlpath.Node{}
	for {
		if !resultsIntr.Next() {
			break
		}

		result := resultsIntr.Node()

		resultWord := pkg.XpathString(result, "word")

		// 입(을) 모으다 -> 밉 모으다
		reExclude := regexp.MustCompile(`\([^)]*\)`)
		excludeWord := reExclude.ReplaceAllString(resultWord, "")

		// 입(을) 모으다 -> 입을 모으다
		reInclude := regexp.MustCompile(`\(|\)`)
		includeWord := reInclude.ReplaceAllString(resultWord, "")

		if item.Hangul != resultWord && item.Hangul != excludeWord && item.Hangul != includeWord {
			continue
		}

		choices = append(choices, result)
		item.Hangul = resultWord
	}

	itemLabel := item.Hangul
	if item.Def.English != "" && len(choices) > 0 {
		itemLabel += " (" + item.Def.English + ")"
	}

	if interactive || len(choices) > 1 {
		fmt.Fprintf(out, "%s -> ", itemLabel)
	}

	var choiceIndex int
	switch len(choices) {
	case 0:
		item.ExternalID = "-"
		if interactive {
			if item.Def.English == "" {
				inBuf := bufio.NewReader(in)
				fmt.Fprintf(out, "Not found.\nEnter custom English definition: ")
				engDef, err := inBuf.ReadString('\n')
				if err != nil {
					return err
				}
				item.Def.English = strings.TrimSpace(engDef)
			} else {
				fmt.Fprintf(out, "%s", item.Def.English)
			}
		}
		return nil
	case 1:
		choiceIndex = 0
		if interactive {
			fmt.Fprintf(out, "%s", pkg.XpathString(choices[choiceIndex], "sense/translation/trans_word"))
		}
	default:
		fmt.Fprintf(out, "Multiple results found:\n")
		for i, choice := range choices {
			fmt.Fprintf(out, " %d) %s\n", i+1, pkg.XpathString(choice, "sense/translation/trans_word"))
		}
		if interactive {
			choiceIndex = promptMultipleChoice(in, out, item, choices)
		} else {
			fmt.Fprintf(out, "Skipping lookup. Set %s option to choose.\n", pkg.OPT_INTERACTIVE)
			return nil
		}
	}

	item.ExternalID = pkg.KrDictID("kor", pkg.XpathString(choices[choiceIndex], "target_code"), "단어")
	return nil
}

func search(q string, options map[string]string) (*xmlpath.Node, error) {
	reExclude := regexp.MustCompile(`\([^)]*\)`)
	q = reExclude.ReplaceAllString(q, "")

	url := fmt.Sprintf(
		"%s/api/search?key=%s&type_search=search&part=word&q=%s&sort=dict&translated=y&trans_lang=1",
		options[pkg.OPT_KRDICT_API_URL],
		options[pkg.OPT_KRDICT_API_KEY],
		q,
	)

	// The searched is a idiom, aka not a word
	if strings.Contains(q, " ") {
		url = fmt.Sprintf(
			"%s/api/search?key=%s&type_search=search&part=ip&q=%s&sort=dict&translated=y&trans_lang=1",
			options[pkg.OPT_KRDICT_API_URL],
			options[pkg.OPT_KRDICT_API_KEY],
			strings.ReplaceAll(q, " ", "+"),
		)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify : true},
	}
	client := &http.Client{Transport: tr}

	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("Non-200 searching KR DICT API")
	}

	return xmlpath.Parse(resp.Body)
}

func promptMultipleChoice(in io.Reader, out io.Writer, item *pkg.Item, choices []*xmlpath.Node) int {
	inBuf := bufio.NewReader(in)
	for {
		fmt.Fprintf(out, "Enter number: ")
		answerString, err := inBuf.ReadString('\n')
		if err != nil {
			fmt.Fprintf(out, "%s\n", err)
			continue
		}

		answerNum, err := strconv.Atoi(strings.TrimSpace(answerString))
		if err != nil || answerNum < 1 || answerNum > len(choices) {
			fmt.Fprintf(out, "Invalid number\n")
			continue
		}

		return answerNum - 1
	}
}
