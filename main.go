package main

import (
	"bayers_spam/tokenize"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func readFileList() map[string][]string {
	out := map[string][]string{}
	out["spam"] = []string{}
	out["ham"] = []string{}

	file, err := os.Open("trec06c/full/index")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	// optionally, resize scanner's capacity for lines over 64K, see next example
	for scanner.Scan() {
		line := scanner.Text()
		// fmt.Println(line)

		elem := strings.Split(line, " ")

		if len(elem) < 2 {
			continue
		}

		if elem[0] == "spam" {
			out["spam"] = append(out["spam"], elem[1])
		} else {
			out["ham"] = append(out["ham"], elem[1])
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	return out
}

func buildDict(files []string) {
	wordCountMap := map[string]int64{}
	for _, file := range files {
		content, err := os.ReadFile("trec06c/full/" + file)
		if err != nil {
			log.Fatal(err)
		}
		contentUTF, _ := GbkToUtf8([]byte(content))
		tokens := tokenize.GetInstance().Analyse(string(contentUTF))

		// f, err := os.Open("trec06c/full/" + file)
		// if err != nil {
		// 	log.Fatal(err)
		// }

		// scanner := bufio.NewScanner(f)
		// // optionally, resize scanner's capacity for lines over 64K, see next example
		// for scanner.Scan() {
		// 	line := scanner.Text()
		// 	lineUTF, _ := GbkToUtf8([]byte(line))
		// 	// fmt.Println(line)
		// 	//FIXME: 这里应该提取文本特征，而不应该只做分词，不然可能会产生大量的分词结果，导致计算一个文本的概率越乘越小
		// 	tokens := tokenize.GetInstance().TokenizeCut(string(lineUTF))
		for _, t := range tokens {
			if _, ok := wordCountMap[t]; ok {
				wordCountMap[t] = wordCountMap[t] + 1
			} else {
				wordCountMap[t] = 1
			}
		}
		// }
		// f.Close()
		// if err := scanner.Err(); err != nil {
		// 	log.Fatal(err)
		// }
	}

	count := 0
	for k, v := range wordCountMap {
		fmt.Printf("%v=%v\n", k, v)
		count++
		if count > 10 {
			break
		}
	}

	wcm, _ := json.Marshal(wordCountMap)
	wordCountMapToFile("wordcount_ham_x.json", wcm)
}

func GbkToUtf8(s []byte) ([]byte, error) {
	reader := transform.NewReader(bytes.NewReader(s), simplifiedchinese.GBK.NewDecoder())
	d, e := ioutil.ReadAll(reader)
	if e != nil {
		return nil, e
	}
	return d, nil
}

func wordCountMapToFile(file string, content []byte) {
	// You can also write it to a file as a whole
	err := os.WriteFile(file, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func fileToWordCountMap(file string) map[string]int64 {
	out := map[string]int64{}
	content, err := os.ReadFile(file)
	if err != nil {
		log.Fatal(err)
	}

	err = json.Unmarshal(content, &out)
	if err != nil {
		log.Fatal(err)
	}
	return out
}

func Calculate() (map[string]float64, map[string]float64) {
	spamWordMap := fileToWordCountMap("wordcount_spam2.json")
	hamWordMap := fileToWordCountMap("wordcount_ham2.json")

	var spamSum int64 = 0
	var hamSum int64 = 0

	for _, v := range spamWordMap {
		spamSum += v
	}

	for _, v := range hamWordMap {
		hamSum += v
	}

	fmt.Printf("spamSum=%v, hamSum=%v\n", spamSum, hamSum)

	spamWord_wc := map[string]float64{}
	hamWord_wc := map[string]float64{}

	for k, v := range spamWordMap {
		spamWord_wc[k] = float64(v+1) / float64(spamSum+int64(len(spamWordMap)))
	}

	for k, v := range hamWordMap {
		hamWord_wc[k] = float64(v+1) / float64(hamSum+int64(len(hamWordMap)))
	}

	return spamWord_wc, hamWord_wc
}

const pc_spam = float64(42854) / float64(42854+21766)
const pc_ham = float64(21766) / float64(42854+21766)

func predict(content string, spamWord_wc, hamWord_wc map[string]float64) (float64, float64) {
	tokens := tokenize.GetInstance().Analyse(content)

	spam_dc := 1.0
	for _, t := range tokens {
		v, ok := spamWord_wc[t]
		if ok {
			spam_dc = spam_dc * v
			fmt.Printf("%s, spam_dc = spam_dc * v = %v\n", t, spam_dc)
		}
	}

	ham_dc := 1.0
	for _, t := range tokens {
		v, ok := hamWord_wc[t]
		if ok {
			ham_dc = ham_dc * v
		}
	}

	fmt.Printf("spam_dc * pc_spam = %v\n", spam_dc*pc_spam)
	fmt.Printf("ham_dc * pc_ham = %v\n", ham_dc*pc_ham)
	fmt.Printf("spam_dc*pc_spam + ham_dc*pc_ham = %v\n", spam_dc*pc_spam+ham_dc*pc_ham)

	spam_predict := spam_dc * pc_spam / (spam_dc*pc_spam + ham_dc*pc_ham)
	ham_predict := ham_dc * pc_ham / (spam_dc*pc_spam + ham_dc*pc_ham)
	return spam_predict, ham_predict
}

//若该词只出现在垃圾邮件的词典中，则令 P(w|s′)=0.01，反之亦然；若都未出现，则令 P(s|w)=0.4。PS.这里做的几个假设基于前人做的一些研究工作得出的。
func main() {
	// fileList := readFileList()
	// fmt.Printf("%v, %v\n", len(fileList["spam"]), len(fileList["ham"]))
	// buildDict(fileList["spam"])
	// buildDict(fileList["ham"])

	// tokenize.GetInstance()
	spamWord_wc, hamWord_wc := Calculate()
	sp, hp := predict(`公司现在有内部推荐机会,2-3人
	主要从事视频编解码器在pc/dsp/arm上的优化工作.
	(h264,mpeg4-part2,wmv)编解码都做.
	音频也在做，需要对aac,mp3比较熟悉的人员．
	希望你有比较丰富的优化经验,
	能认真钻研技术,人品好.
	我们需要作出好的产品来,
	不只是纸上谈兵而已.
	欢迎你能尽快到公司工作．
	暂不招收学生，条件优秀的我们可以保持联系．
	工作地点：上海
	简历请寄至：bshymq_wang@yahoo.com.cn
	谢谢！来信必复．`, spamWord_wc, hamWord_wc)
	fmt.Printf("Bayers, spam_prob=%v, ham_prob=%v", sp, hp)
}
