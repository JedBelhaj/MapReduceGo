readme_content = """

# Projet MapReduce Distribué

## Présentation

Ce projet implémente un système MapReduce distribué en Go avec un tableau de bord web pour suivre en temps réel l’avancement des tâches et l’état des workers.

Le système se compose d’un processus **master** qui coordonne les tâches map et reduce, et de plusieurs processus **worker** qui exécutent les tâches. Le master sert également une interface web pour la supervision.

---

## Structure du projet

- `main.go` : point d’entrée pour lancer le master ou un worker.
- `mapreduce/` : logique principale de MapReduce, incluant les implémentations master et worker.
- `mapreduce/dashboard.html` : page HTML du tableau de bord.
- `...` : autres fichiers sources du projet.

---

## Lancer le projet

### Prérequis

- Go 1.18+ installé.
- Terminal ou shell.

---

### Compiler le projet

```bash
go build -o mapreduce main.go
```
