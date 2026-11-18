package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"
)

type Nom struct {
	Name string `json:"name"`
}

type Type struct {
	Type Nom `json:"type"`
}

type Pokemon struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Types []Type `json:"types"`
}

type ListePokemon struct {
	Results []Pokemon `json:"results"`
}

type Traduction struct {
	Name     string `json:"name"`
	Language Nom    `json:"language"`
}

type Espece struct {
	Names []Traduction `json:"names"`
}

type PagePokemon struct {
	Pokemon Pokemon
	NomFr   string
}

var client = &http.Client{Timeout: 5 * time.Second}

func appelerAPI(adresse string, resultat any) error {
	resp, err := client.Get("https://pokeapi.co/api/v2/" + adresse)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 404 {
		return nil
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("l'API a répondu %d", resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(resultat)
}

func getPokemons() ([]Pokemon, error) {
	var liste ListePokemon
	err := appelerAPI("pokemon?limit=151", &liste)
	for i := range liste.Results {
		liste.Results[i].ID = i + 1
	}
	return liste.Results, err
}

func getPokemon(id string) (Pokemon, error) {
	var pokemon Pokemon
	err := appelerAPI("pokemon/"+id, &pokemon)
	return pokemon, err
}

func getEspece(id string) (Espece, error) {
	var espece Espece
	err := appelerAPI("pokemon-species/"+id, &espece)
	return espece, err
}

func nomFrancais(espece Espece) string {
	for _, traduction := range espece.Names {
		if traduction.Language.Name == "fr" {
			return traduction.Name
		}
	}
	return ""
}

func erreur(w http.ResponseWriter, code int, message string) {
	w.WriteHeader(code)
	tmpl := template.Must(template.ParseFiles("templates/erreur.html"))
	tmpl.Execute(w, message)
}

func apiEnPanne(w http.ResponseWriter, err error) {
	fmt.Println("Erreur :", err)
	erreur(w, 500, "Le Pokédex ne répond pas. Réessayez plus tard.")
}

func accueil(w http.ResponseWriter, r *http.Request) {
	pokemons, err := getPokemons()
	if err != nil {
		apiEnPanne(w, err)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/index.html"))
	tmpl.Execute(w, pokemons)
}

func fiche(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	pokemon, err := getPokemon(id)
	if err != nil {
		apiEnPanne(w, err)
		return
	}
	if pokemon.ID == 0 {
		erreur(w, 404, "Ce Pokémon n'existe pas.")
		return
	}

	espece, err := getEspece(id)
	if err != nil {
		apiEnPanne(w, err)
		return
	}

	page := PagePokemon{
		Pokemon: pokemon,
		NomFr:   nomFrancais(espece),
	}

	tmpl := template.Must(template.ParseFiles("templates/fiche.html"))
	tmpl.Execute(w, page)
}

func introuvable(w http.ResponseWriter, r *http.Request) {
	erreur(w, 404, "Cette page n'existe pas.")
}

func main() {
	http.HandleFunc("/{$}", accueil)
	http.HandleFunc("/pokemon/{id}", fiche)
	http.HandleFunc("/", introuvable)
	http.Handle("/static/", http.FileServer(http.Dir(".")))

	fmt.Println("Serveur lancé sur http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
