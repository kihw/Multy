# Utilise une image Windows avec Go
FROM golang:1.18-windowsservercore

# Installe Chocolatey et gcc via mingw
RUN powershell -Command \
    Set-ExecutionPolicy Bypass -Scope Process -Force; \
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; \
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1')); \
    choco install mingw -y

# Ajoute gcc au PATH
RUN setx PATH "%PATH%;C:\Program Files\mingw64\bin"

# Définit le répertoire de travail
WORKDIR /app

# Copie les fichiers de dépendances locaux et l'application dans le conteneur
COPY . .
COPY vendor/ ./vendor/

# Compile l'application avec les dépendances locales
RUN go build -mod=vendor -o main.exe .

# Expose le port de l'application
EXPOSE 8080

# Commande pour exécuter l'application
CMD ["./main.exe"]
