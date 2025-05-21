package mapreduce

import (
	"encoding/json"
	"hash/fnv"
	"io/ioutil"
	"log"
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
	inFile = inFile + ".txt"
	content, err := ioutil.ReadFile(inFile)
	if err != nil {
		log.Fatalf("DoMap: cannot read file %v", err)
	}

	kvs := mapF(string(content))

	// Create encoders for each reduce file
	files := make([]*os.File, nReduce)
	encoders := make([]*json.Encoder, nReduce)
	for i := 0; i < nReduce; i++ {
		fileName := ReduceName(jobName, mapTaskNumber, i)
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("DoMap: cannot create file %s: %v", fileName, err)
		}
		files[i] = file
		encoders[i] = json.NewEncoder(file)
	}

	for _, kv := range kvs {
		r := int(ihash(kv.Key)) % nReduce
		err := encoders[r].Encode(&kv)
		if err != nil {
			log.Fatalf("DoMap: encode error: %v", err)
		}
	}

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
	keyGroups := make(map[string][]string)

	// Read intermediate files
	for i := 0; i < nMap; i++ {
		fileName := ReduceName(jobName, i, reduceTaskNumber)
		file, err := os.Open(fileName)
		if err != nil {
			log.Fatalf("DoReduce: cannot open %s: %v", fileName, err)
		}
		decoder := json.NewDecoder(file)
		var kv KeyValue
		for decoder.Decode(&kv) == nil {
			keyGroups[kv.Key] = append(keyGroups[kv.Key], kv.Value)
		}
		file.Close()
	}

	// Open output file
	outFileName := MergeName(jobName, reduceTaskNumber)
	outFile, err := os.Create(outFileName)
	if err != nil {
		log.Fatalf("DoReduce: cannot create %s: %v", outFileName, err)
	}
	encoder := json.NewEncoder(outFile)

	// Sort keys for deterministic output
	var keys []string
	for k := range keyGroups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		result := reduceF(k, keyGroups[k])
		encoder.Encode(&KeyValue{Key: k, Value: result})
	}
	outFile.Close()
}
