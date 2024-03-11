# Go Rate Limiter
Este projeto é um `rate limiter` implementado em Go. Ele usa um banco de dados Redis para armazenar os dados do cliente e limitar as requisições com base no IP ou no token do cliente.

### Configuração do Projeto
1. Clone o repositório para a sua máquina local usando git clone.
2. Navegue até o diretório do projeto.
3. Instale as dependências do projeto com o comando go mod download.

### Executando o Projeto
Para executar o projeto, você pode usar o comando `go run` no diretório `cmd/server` (é necessário que o container do redis esteja em execução):

```sh
$ go run cmd/server/main.go
```

O servidor estará rodando na porta 8080.

### Executando os Testes

Para executar os testes, você pode usar o comando `go test` no diretório `pkg/ratelimiter`:

```sh
$ go test ./pkg/ratelimiter
```

### Docker
Este projeto inclui um arquivo `docker-compose.yml` que pode ser usado para iniciar um servidor Redis. Para iniciar o servidor Redis, use o comando `docker-compose up`, ou `docker compose up` dependendo da versão do docker compose que estiver executando em sua máquina:

```sh
$ docker-compose up | docker compose up
```

### Makefile
Um Makefile também está incluído para simplificar algumas tarefas comuns. Para ver as tarefas disponíveis, use o comando make help.

Entre os comandos, é possível executar testes de carga através da ferramenta `siege` (garanta que está instalada em sua máquina).

Os testes possíveis contemplam ambos os cenários: bloqueio por IP e por Token, sendo respectivamente:

```sh
# teste de carga para garantir o bloqueio por IP
$ make test

# teste de carga para garantir o bloqueio por token
$ make test-api-key
```