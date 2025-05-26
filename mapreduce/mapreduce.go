package mapreduce

import (
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

const prefix = "mrtmp."

// KeyValue is a type used to hold the key/value pairs passed to the map and
// reduce functions.
type KeyValue struct {
	Key   string
	Value string
}

// reduceName constructs the name of the intermediate file which map task
// <mapTask> produces for reduce task <reduceTask>.
func ReduceName(jobName string, mapTask int, reduceTask int) string {
	return prefix + jobName + "-" + strconv.Itoa(mapTask) + "-" + strconv.Itoa(reduceTask)
}

// mergeName constructs the name of the output file of reduce task <reduceTask>
func MergeName(jobName string, reduceTask int) string {
	return prefix + jobName + "-res-" + strconv.Itoa(reduceTask)
}

// ansName constructs the name of the output file of the final answer
func AnsName(jobName string) string {
	return prefix + jobName
}

// clean all intermediary files generated for a job
func CleanIntermediary(jobName string, nMap, nReduce int) {
	// Supprimer les fichiers intermédiaires produits les tâches map
	for reduceTNbr := 0; reduceTNbr < nReduce; reduceTNbr++ {
		for mapTNbr := 0; mapTNbr < nMap; mapTNbr++ {
			os.Remove(ReduceName(jobName, mapTNbr, reduceTNbr))
		}
		os.Remove(MergeName(jobName, reduceTNbr))
	}
}

// Is used to associate to each key a unique reduce file
func ihash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

// doMap applique la fonction mapF, et sauvegarde les résultats.
// A COMPLETER
func DoMap(
	jobName string,
	mapTaskNumber int,
	inFile string,
	nReduce int,
	mapF func(contents string) []KeyValue,
) {
	// Lire le contenu du fichier d’entrée
	content, err := ioutil.ReadFile(inFile)
	if err != nil {
		panic("Erreur lecture fichier d'entrée: " + err.Error())
	}

	// Appliquer mapF pour obtenir les paires clé/valeur
	kvs := mapF(string(content))

	// Créer un tableau d'encodeurs JSON, un par fichier reduce
	encoders := make([]*json.Encoder, nReduce)
	files := make([]*os.File, nReduce)

	// Créer et ouvrir les fichiers intermédiaires pour chaque reduce
	for r := 0; r < nReduce; r++ {
		fileName := ReduceName(jobName, mapTaskNumber, r)
		file, err := os.Create(fileName)
		if err != nil {
			panic("Erreur création fichier reduce: " + err.Error())
		}
		files[r] = file
		encoders[r] = json.NewEncoder(file)
	}

	// Pour chaque paire, calculer le reduceTask correspondant et l’écrire
	for _, kv := range kvs {
		r := int(ihash(kv.Key)) % nReduce
		err := encoders[r].Encode(&kv)
		if err != nil {
			panic("Erreur écriture kv dans fichier intermédiaire: " + err.Error())
		}
	}

	// Fermer les fichiers
	for _, f := range files {
		f.Close()
	}
}

// doReduce effectue une tâche de réduction en lisant les fichiers
// intermédiaires, en regroupant les valeurs par clé, et en appliquant
// la fonction reduceF.
// A COMPLETER
func DoReduce(
	jobName string,
	reduceTaskNumber int,
	nMap int,
	reduceF func(key string, values []string) string,
) {
	// Map de regroupement des valeurs par clé
	keyValues := make(map[string][]string)

	// Lire chaque fichier intermédiaire produit par les tâches Map
	for i := 0; i < nMap; i++ {
		fileName := ReduceName(jobName, i, reduceTaskNumber)
		// Ouvrir le fichier pour la tâche de mappage i
		file, err := os.Open(fileName)
		if err != nil {
			panic("Erreur ouverture fichier intermédiaire: " + err.Error())
		}
		// Lire les paires clé-valeur du fichier
		// Ajouter la valeur à la liste de valeurs pour cette clé
		decoder := json.NewDecoder(file)
		var kv KeyValue
		for decoder.Decode(&kv) == nil {
			keyValues[kv.Key] = append(keyValues[kv.Key], kv.Value)
		}
		file.Close()
	}

	// Récupérer les clés triées pour un ordre déterministe
	var keys []string
	for k := range keyValues {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Ouvrir le fichier de sortie pour la tâche de réduction
	// utiliser MergeName
	outputFile, err := os.Create(MergeName(jobName, reduceTaskNumber))
	if err != nil {
		panic("Erreur création fichier résultat reduce: " + err.Error())
	}
	defer outputFile.Close()

	// Créer un encodeur JSON pour le fichier de sortie
	enc := json.NewEncoder(outputFile)

	// Réduire chaque clé et écrire le résultat
	// Appliquer la fonction de réduction à chaque clé
	for _, key := range keys {
		// Appliquer reduceF pour réduire les valeurs associées à cette clé
		// Écrire la clé et la valeur réduite dans le fichier de sortie
		reducedValue := reduceF(key, keyValues[key])
		err := enc.Encode(&KeyValue{Key: key, Value: reducedValue})
		if err != nil {
			panic("Erreur encodage résultat reduce: " + err.Error())
		}
	}
}

// concatFiles concatène plusieurs fichiers en un seul
func Sequential(jobName string, files []string, nReduce int, mapF func(string) []KeyValue, reduceF func(string, []string) string) {
	for i, file := range files {
		DoMap(jobName, i, file, nReduce, mapF)
	}

	for i := 0; i < nReduce; i++ {
		DoReduce(jobName, i, len(files), reduceF)
	}

	// Merge results
	resFiles := make([]string, 0, nReduce)
	for i := 0; i < nReduce; i++ {
		resFiles = append(resFiles, MergeName(jobName, i))
	}
	err := concatFiles(AnsName(jobName), resFiles)
	CheckError(err, "cannot merge output files: %v\n")
}
