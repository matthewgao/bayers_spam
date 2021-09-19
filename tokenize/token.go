package tokenize

import "github.com/yanyiwu/gojieba"

var mgr *TokenizerMgr

type TokenizerMgr struct {
	jieba *gojieba.Jieba
}

func GetInstance() *TokenizerMgr {
	if mgr == nil {
		userDict := gojieba.USER_DICT_PATH

		x := gojieba.NewJieba(
			gojieba.DICT_PATH,
			gojieba.HMM_PATH,
			userDict,
			gojieba.IDF_PATH,
			gojieba.STOP_WORDS_PATH)

		mgr = &TokenizerMgr{
			jieba: x,
		}
	}

	return mgr
}

func (this *TokenizerMgr) TokenizeCutAll(s string) []string {
	return this.jieba.CutAll(s)
}

func (this *TokenizerMgr) TokenizeCut(s string) []string {
	return this.jieba.Cut(s, true)
}

func (this *TokenizerMgr) Analyse(s string) []string {
	return this.jieba.Extract(s, 20)
}
