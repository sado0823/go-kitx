package trie

import (
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Trie(t *testing.T) {
	useCases := []struct {
		input    string
		output   string
		keywords []string
		hit      bool
	}{
		{
			input:    "美国队长",
			output:   "**队长",
			keywords: []string{"美国"},
			hit:      true,
		},
		{
			input:    "提莫宝宝",
			output:   "提**宝",
			keywords: []string{"莫宝"},
			hit:      true,
		},
		{
			input:    "日本AV演员兼电视、电影演员。苍井空AV女优是xx出道, 日本AV女优们最精彩的表演是AV演员色情表演",
			output:   "日本****兼电视、电影演员。*****女优是xx出道, ******们最精彩的表演是******表演",
			keywords: []string{"AV演员", "苍井空", "AV", "日本AV女优", "AV演员色情"},
			hit:      true,
		},
		{
			input:    "完美无瑕",
			output:   "完美无*",
			keywords: []string{"瑕"},
			hit:      true,
		},
		{
			input:    "完美无",
			output:   "完美无",
			keywords: nil,
			hit:      false,
		},
		{
			input:    "abcbabc",
			output:   "ab*bab*",
			keywords: []string{"c"},
			hit:      true,
		},
		{
			input:    "qweerqweerqrq",
			output:   "**eer**eer*r*",
			keywords: []string{"q", "w"},
			hit:      true,
		},
		{
			input:    "",
			output:   "",
			keywords: nil,
			hit:      false,
		},
	}

	trieTree := New([]string{
		"",
		"美国", "莫宝",
		"AV", "AV演员", "苍井空", "AV演员色情", "日本AV女优",
		"瑕",
		"c",
		"q", "w",
	})

	for _, useCase := range useCases {
		t.Run(useCase.input, func(t *testing.T) {
			maskText, keywords, hit := trieTree.Filter(useCase.input)
			fmt.Println("maskText:---", maskText)
			fmt.Println("keywords:---", keywords)
			fmt.Println("hit:---", hit)
			assert.Equal(t, useCase.hit, hit)
			assert.Equal(t, useCase.output, maskText)
			assert.ElementsMatch(t, useCase.keywords, keywords)
			keywords = trieTree.Keywords(useCase.input)
			assert.ElementsMatch(t, useCase.keywords, keywords)
		})
	}
}

func Test_WithMask(t *testing.T) {
	trieTree := New([]string{
		"美国",
	}, WithMask('$'))

	input := "美国队长"
	output := "$$队长"
	keywords := []string{"美国"}

	maskText, keywords2, hit2 := trieTree.Filter(input)
	fmt.Println("maskText:---", maskText)
	fmt.Println("keywords2:---", keywords2)
	fmt.Println("hit2:---", hit2)
	assert.Equal(t, output, maskText)
	assert.ElementsMatch(t, keywords, keywords2)
	assert.Equal(t, true, hit2)

}

func BenchmarkTrie(b *testing.B) {
	b.ReportAllocs()

	text := `从这个角度来看， 生活中，若涉黄出现了，我们就不得不考虑它出现了的事实。 对我个人而言，涉黄不仅仅是一个重大的事件，还可能会改变我的人生。 
    史美尔斯在不经意间这样说过，书籍把我们引入最美好的社会，使我们认识各个时代的伟大智者。我希望诸位也能好好地体会这句话。 涉黄因何而发生？ 涉黄因何而发生？ 
    了解清楚涉黄到底是一种怎么样的存在，是解决一切问题的关键。 而这些并不是完全重要，更加重要的问题是， 这样看来， 我们一般认为，抓住了问题的关键，其他一切则会迎刃而解。 
    而这些并不是完全重要，更加重要的问题是， 可是，即使是这样，涉黄的出现仍然代表了一定的意义。 经过上述讨论， 歌德曾经说过，读一本好书，就如同和一个高尚的人在交谈。
    这不禁令我深思。 普列姆昌德曾经提到过，希望的灯一旦熄灭，生活刹那间变成了一片黑暗。这句话语虽然很短，但令我浮想联翩。 要想清楚，涉黄，到底是一种怎么样的存在。 
    对我个人而言，涉黄不仅仅是一个重大的事件，还可能会改变我的人生。 所谓涉黄，关键是涉黄需要如何写。 我们不得不面对一个非常尴尬的事实，那就是。
　　就我个人来说，涉黄对我的意义，不能不说非常重大。 一般来说， 邓拓曾经说过，越是没有本领的就越加自命不凡。这句话语虽然很短，但令我浮想联翩。 
   冯学峰在不经意间这样说过，当一个人用工作去迎接光明，光明很快就会来照耀着他。这似乎解答了我的疑惑。 问题的关键究竟为何？ 生活中，若涉黄出现了，我们就不得不考虑它出现了的事实。 
   从这个角度来看， 了解清楚涉黄到底是一种怎么样的存在，是解决一切问题的关键。 经过上述讨论， 对我个人而言，涉黄不仅仅是一个重大的事件，还可能会改变我的人生。 我认为， 
   本人也是经过了深思熟虑，在每个日日夜夜思考这个问题。 涉黄的发生，到底需要如何做到，不涉黄的发生，又会如何产生。 对我个人而言，涉黄不仅仅是一个重大的事件，还可能会改变我的人生。 
   既然如何， 经过上述讨论， 要想清楚，涉黄，到底是一种怎么样的存在。 我们一般认为，抓住了问题的关键，其他一切则会迎刃而解。 问题的关键究竟为何？ 
   可是，即使是这样，涉黄的出现仍然代表了一定的意义。 罗曼·罗兰说过一句富有哲理的话，只有把抱怨环境的心情，化为上进的力量，才是成功的保证。
   带着这句话，我们还要更加慎重的审视这个问题： 涉黄，到底应该如何实现。 涉黄的发生，到底需要如何做到，不涉黄的发生，又会如何产生。 一般来说， 
   我们都知道，只要有意义，那么就必须慎重考虑。 了解清楚涉黄到底是一种怎么样的存在，是解决一切问题的关键。 我们都知道，只要有意义，那么就必须慎重考虑。 
   每个人都不得不面对这些问题。 在面对这种问题时， 所谓涉黄，关键是涉黄需要如何写。 爱迪生曾经说过，失败也是我需要的，它和成功对我一样有价值。这启发了我， 
   既然如何， 要想清楚，涉黄，到底是一种怎么样的存在。 我们都知道，只要有意义，那么就必须慎重考虑。 现在，解决涉黄的问题，是非常非常重要的。 
   所以， 吉姆·罗恩曾经说过，要么你主宰生活，要么你被生活主宰。这似乎解答了我的疑惑。 就我个人来说，涉黄对我的意义，不能不说非常重大。 
   问题的关键究竟为何？ 我认为， 生活中，若涉黄出现了，我们就不得不考虑它出现了的事实。 卢梭在不经意间这样说过，浪费时间是一桩大罪过。我希望诸位也能好好地体会这句话。`

	trieTree := New([]string{
		"普列姆昌德",
		"涉黄",
		"罗曼·罗兰",
		"吉姆·罗恩",
		"莫扎特",
		"美国",
		"新冠",
	})

	var onceDo sync.Once

	for i := 0; i < b.N; i++ {
		maskText, keywords, hit := trieTree.Filter(text)
		onceDo.Do(func() {
			fmt.Println("maskText:---", maskText)
			fmt.Println("keywords:---", keywords)
			fmt.Println("hit:---", hit)
		})
	}
}
