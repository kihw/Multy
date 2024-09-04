# Utilise une image Go officielle
FROM golang:1.23

# Définit le répertoire de travail dans le conteneur
WORKDIR /app

# Copie les fichiers du module Go et le fichier `go.sum` dans le répertoire de travail
COPY go.mod go.sum ./

# Télécharge les dépendances du projet
RUN go mod download
# Copie tout le reste du projet dans le répertoire de travail
COPY . .

# Compile l'application
RUN go build -o /app/main

# Définit le port sur lequel l'application sera exposée
EXPOSE 8080

# Définit la commande pour démarrer l'application
CMD ["/app/main"]
