package main

import (
	"math/rand"
	"strings"
)

var verboSemObj = []string{
	"trollou",
	"feedou",
	"farmou",
	"fodeu",
}

var verboClanUva = []string{
	"matar",
	"mutar",
}

var moios = []string{
	"moios",
	"broios",
	"troios",
	"droios",
	"trodos",
	"trotos",
	"badaras",
	"baradas",
}

var clanUva = []string{
	"o Adriano",
	"o Adriano",
	"o Adriano",
	"o Adriano",
	"o Adriano",
	"o Kibe",
	"o Pastuch",
	"o Caio",
	"o Sebben",
	"o Pedro",
	"o Dezzo",
}

var especial = []string{
	"Ó",
	"EU SOU O LUQUITOOO AHHHH",
	"tá na jungle farmando hard",
	"vou aplicar a lei do solinho",
	"que se foda catequese",
	"pai tá chato",
	"pai tá chato, né?",
	"dalhe na narguilheira",
	"o cara é um mamute",
	"o cara é um mamute, né?",
	"ADURIANO NEH",
	"Boneco de posto",
	"tá maluco, tá doidão",
	"piá, é isso piá",
	"tem que ver isso aí",
	"antes tide do que kunkka",
	"ADURIANO NÉ",
}

var adjetivoDmais = []string{
	"gordo",
	"carinhas",
	"gado",
	"badaras",
	"baradas",
	"pipinhas léguas",
}

var adjetivoDo = []string{
	"baby",
	"xesk",
	"bresk",
	"chesque",
	"bait",
}

var adicionalFinal = []string{
	"papai",
	"meu papai",
	"pai",
	"meu pai",
	"primo",
	"meu primo",
	"papebas",
	"meu papebas",
	"tá ligado",
}

var adicionalInicio = []string{
	"piá",
}

var jogar = []string{
	"feedar",
	"dar uma rabiada",
}

var jogos = []string{
	"Dota 2 (Lion mid)",
	"Dota 2 (Pudge mid)",
	"Dota 2 (Legion jungle)",
	"Dota 2 (Magnus mid)",
	"Poketibia",
	"CoD",
	"Grand Chase",
	"Tibia",
	"Fallguebas",
	"Minhocas",
	"League of Legends (Brand jungle)",
}

var erroAudioJaTocando = []string{
	"Uma coisa de cada vez",
	"Só posso tocar um áudio",
}

func frasePiorQue() string {
	r := rand.Intn(len(moios))
	s := moios[r]
	return "Pior que daí é " + s
}

func fraseAi1() string {
	r := rand.Intn(len(verboSemObj))
	s := verboSemObj[r]
	return "Aí " + s
}

func fraseAi2() string {
	r := rand.Intn(len(moios))
	s := moios[r]
	return "Aí é " + s
}

func fraseEuVou() string {
	r1 := rand.Intn(len(verboClanUva))
	r2 := rand.Intn(len(clanUva))
	s1 := verboClanUva[r1]
	s2 := clanUva[r2]
	return "Eu vou " + s1 + " " + s2
}

func fraseDmais() string {
	r := rand.Intn(len(adjetivoDmais))
	s := adjetivoDmais[r]
	s = s + " d+"
	return fraseAdicionalAmbas(s)
}

func fraseEspecial() string {
	r := rand.Intn(len(especial))
	s := especial[r]
	return s
}

func fraseDo() string {
	r := rand.Intn(len(adjetivoDo))
	s := adjetivoDo[r]
	s = s + " do " + s
	return fraseAdicionalAmbas(s)
}

func fraseAdicionalFinal(s string) string {
	r := rand.Intn(100)
	if r <= 50 {
		return s
	}

	r = rand.Intn(len(adicionalFinal))
	return s + ", " + adicionalFinal[r]
}

func fraseAdicionalInicial(s string) string {
	r := rand.Intn(100)
	if r <= 50 {
		return s
	}

	r = rand.Intn(len(adicionalInicio))
	return adicionalInicio[r] + ", " + s
}

func fraseAdicionalAmbas(s string) string {
	return fraseAdicionalFinal(fraseAdicionalInicial(s))
}

func frase() string {
	r := rand.Intn(140)
	if r <= 20 {
		return fraseAi1()
	} else if r >= 21 && r <= 40 {
		return fraseAi2()
	} else if r >= 41 && r <= 60 {
		return fraseEuVou()
	} else if r >= 61 && r <= 80 {
		return fraseDmais()
	} else if r >= 81 && r <= 100 {
		return fraseDo()
	} else if r >= 101 && r <= 120 {
		return frasePiorQue()
	} else if r >= 121 && r <= 140 {
		return fraseEspecial()
	}
	return ""
}

// GeraFrase gera uma frase aleatória
func GeraFrase() string {
	s := frase()
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}

// GeraJogo gera um jogo aleatório
func GeraJogo() string {
	r := rand.Intn(len(jogos))
	s := jogos[r]
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}

// GeraErroAudioJaTocando gera uma mensagem de erro
func GeraErroAudioJaTocando() string {
	r := rand.Intn(len(erroAudioJaTocando))
	s := erroAudioJaTocando[r]
	s = fraseAdicionalFinal(s)
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}
