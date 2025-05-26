# Système MapReduce Distribué avec Dashboard Web

## Description du projet

Ce projet implémente un système distribué de type MapReduce en Go, incluant un tableau de bord web pour le monitoring en temps réel des tâches et des workers. Le système permet de traiter efficacement de grands volumes de données en parallélisant les calculs sur plusieurs machines.

L'implémentation comprend :
- Un master qui coordonne l'ensemble du processus
- Des workers qui exécutent les tâches map ou reduce
- Un dashboard web pour visualiser l'état du système en temps réel

## Structure des fichiers

```
PDIST25-AYMENJEDDOU-ABDERRAHIMNEJI-MOHAMEDAMINETRABELSI/
├── cmd/
│   ├── master/
│   │   └── main.go       # Point d'entrée pour le master
│   └── worker/
│       └── main.go       # Point d'entrée pour le worker
├── input/ 
    ├──input1.txt
    ├──input2.txt
    ├──input3large.txt                   # Fichiers d'entrée pour les tests(le troisieme est trés long)
├── mapreduce/
│   ├── common.go
│   ├── mapreduce.go      # Implémentation des fonctions DoMap et DoReduce...
│   ├── master.go         # Implémentation du master
│   ├── worker.go         # Implémentation du worker
│   └── word_count.go     # Fonctions spécifiques pour le comptage de mots
├── tests/                # Tests unitaires
└── web/
    └── index.html        # Interface web du dashboard
```

## Prérequis

- Go 1.16 ou supérieur

## Installation

. Accédez au répertoire du projet :
```
cd ../PDIST25-AYMENJEDDOU-ABDERRAHIMNEJI-MOHAMEDAMINETRABELSI 
```

2. Téléchargez les dépendances :
```
go mod download
```

## Compilation

Pour compiler le master et les workers :
(vous ne devez pas le faire par ce que c'est déjà fait)

```
go build -o master.exe ./cmd/master
go build -o worker.exe ./cmd/worker
```

## Exécution

1. Lancez le master :
```
.\master.exe -job testjob -files input/input3large.txt (ou input1.txt ou input2.txt)
```

2. Lancez un ou plusieurs workers (dans des terminaux séparés) :
```
./worker.exe -master localhost:1234 -id worker1
./worker.exe -master localhost:1234 -id worker2
./worker.exe -master localhost:1234 -id worker3
```

3. Accédez au dashboard web via votre navigateur :
```
http://localhost:8080
```

## Fonctionnalités

- **Traitement distribué** : Le système répartit les tâches de mappage et de réduction sur plusieurs workers.
- **Tolérance aux pannes** : Le master détecte les workers lents ou en panne et réattribue leurs tâches.
- **Monitoring en temps réel** : Le dashboard web affiche l'état des tâches et des workers.
- **Simulation de pannes** : Les workers peuvent simuler des crashs ou des retards pour tester la robustesse du système.

##Visualisation des résultats

les résultats du countwords seront dans le fichier mrtmp.testjob

get-Content mrtmp.testjob  #commande dans le terminal du master
## Tests

Pour exécuter les tests unitaires :

```
go test ./tests
```

## Réalisé par :

- Aymen Jeddou
- Abderrahim Neji
- Mohamed Amine Trabelsi

Faculté des Sciences de Bizerte, 1ère année cycle ingénieur, 2025
