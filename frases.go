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
	"mamar",
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
	"o Arthur",
	"o Tadeu",
	"o Wandley",
	"o Frost",
	"o Josias",
	"o Voraz",
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
	"Boneco de posto",
	"tá maluco, tá doidão",
	"piá, é isso piá",
	"tem que ver isso aí",
	"antes tide do que kunkka",
	"ADURIANO NEH",
	"ADURIANO NÉ",
	"Tem que viver na poligamia",
	"deu aulaxx",
	"tchama tchama",
}

var adjetivoDmais = []string{
	"gordo",
	"carinhas",
	"gado",
	"badaras",
	"baradas",
	"pipinhas léguas",
}

var palavraDo = []string{
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
	"feedar no",
	"dar uma feedada no",
	"jogar",
	"dar uma rabiada no",
	"dar aulaxx no",
}

var jogos = []string{
	"Dota 2 (Lion mid)",
	"Dota 2 (Pudge mid)",
	"Dota 2 (Legion jungle)",
	"Dota 2 (Magnus mid)",
	"Dota 2 (Invoker mid)",
	"Dota 2 (Sniper mid)",
	"Dota 2 (Meme Hammer mid)",
	"Dota 2 (OD mid)",
	"Warcraft III",
	"Poketibia",
	"CSGO",
	"Grand Chase",
	"Tibia",
	"Fallguebas",
	"Minhocas",
	"Worms",
	"League of Legends (Brand jungle)",
	"League of Legends (Udyr suporte)",
	"World of Warcraft",
	"Warcraft III: Refunded",
	"Warcraft III: Reembolsado",
	"Valorant",
	"Badarant",
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

func fraseEuVouJogar() string {
	r1 := rand.Intn(len(jogar))
	r2 := rand.Intn(len(jogos))
	s1 := jogar[r1]
	s2 := jogos[r2]
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
	r := rand.Intn(len(palavraDo))
	s := palavraDo[r]
	s = s + " do " + s
	return fraseAdicionalAmbas(s)
}

func fraseAdicionalFinal(s string) string {
	r := rand.Intn(1)
	if r == 1 {
		return s
	}

	r = rand.Intn(len(adicionalFinal))
	return s + ", " + adicionalFinal[r]
}

func fraseAdicionalInicial(s string) string {
	r := rand.Intn(1)
	if r == 1 {
		return s
	}

	r = rand.Intn(len(adicionalInicio))
	return adicionalInicio[r] + ", " + s
}

func fraseAdicionalAmbas(s string) string {
	return fraseAdicionalFinal(fraseAdicionalInicial(s))
}

func frase() string {
	r := rand.Intn(7)

	switch r {
	case 0:
		return fraseAi1()
	case 1:
		return fraseAi2()
	case 2:
		return fraseEuVou()
	case 3:
		return fraseDmais()
	case 4:
		return fraseDo()
	case 5:
		return frasePiorQue()
	case 6:
		return fraseEspecial()
	case 7:
		return fraseEuVouJogar()
	default:
		return ""
	}
}

func geraFrase() string {
	s := frase()
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}

func geraJogo() string {
	r := rand.Intn(len(jogos))
	s := jogos[r]
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}

func geraErroAudioJaTocando() string {
	r := rand.Intn(len(erroAudioJaTocando))
	s := erroAudioJaTocando[r]
	s = fraseAdicionalFinal(s)
	s = strings.ToUpper(string(s[0])) + s[1:]
	return s
}
