# Sistema de Clima com Tracing Distribuído (Serviço A e Serviço B)

Este projeto implementa um sistema distribuído em Go que recebe um CEP, consulta a cidade associada a esse CEP e retorna as informações climáticas (temperatura em Celsius, Fahrenheit e Kelvin) utilizando dois serviços separados:

- **Serviço A**: Valida o CEP e chama o Serviço B.
- **Serviço B**: Consulta o clima e a cidade com base no CEP.

Além disso, foi implementado **tracing distribuído** com **OpenTelemetry (OTEL)** e **Zipkin** para monitoramento e rastreamento das requisições entre os serviços.

## Requisitos

- **Docker** e **Docker Compose** para orquestração dos containers.
- **Go 1.21** ou superior para rodar os serviços localmente.

## Estrutura do Projeto

- **Serviço A** (porta 8081): Valida o CEP e chama o **Serviço B**.
- **Serviço B** (porta 8082): Faz a consulta da cidade e do clima para o CEP fornecido.
- **Zipkin** (porta 9411): Coletor de traces que armazena e exibe o rastreamento das requisições entre os serviços.

## Como Rodar o Projeto

### Passo 1: Clonar o Repositório

Clone o repositório para o seu ambiente local:

```bash
git clone https://github.com/4lexRossi/weather-opel-go.git
cd weather-opel-go
```
### Passo 2: Instalar Dependências

Certifique-se de que o Go está instalado em sua máquina. Para instalar as dependências do projeto, navegue até as pastas de Serviço A e Serviço B e execute:

```bash
cd service_A
go mod tidy

cd ../service_B
go mod tidy

```

### Passo 3: Configurar o Docker Compose

Na raiz do projeto, temos um arquivo `docker-compose.yml` que configura os serviços e o Zipkin. Para garantir que os containers sejam criados corretamente, execute:
```bash
docker-compose up --build
```
Isso irá:

Construir as imagens Docker para Serviço A e Serviço B.
Subir o container do Zipkin.bui
Expor os serviços nas portas configuradas (8081 para o Serviço A, 8082 para o Serviço B e 9411 para o Zipkin).

### Passo 4: Testar os Serviços
Com todos os containers rodando, agora você pode testar os serviços.

Serviço A (Validação e Redirecionamento)
Faça uma requisição `POST` para o Serviço A com um JSON contendo o CEP:

```bash
curl -X POST http://localhost:8081/cep -d 'v' -H "Content-Type: application/json"
```

Se o CEP for válido, o Serviço A chamará o Serviço B e retornará a resposta com a cidade e as temperaturas.
Se o CEP for inválido, retornará um erro com o código HTTP 422.
Visualizando os Traces no Zipkin
Para monitorar o tracing distribuído, acesse o Zipkin na URL:

http://localhost:9411

Aqui você pode visualizar os traces gerados pelas requisições feitas entre os Serviços A e B.

### Passo 5: Exemplo de Resposta

Ao chamar o Serviço A com um CEP válido (`29902555`), você deverá obter a resposta:

```bash
{
    "city": "São Paulo",
    "temp_C": 28.5,
    "temp_F": 83.3,
    "temp_K": 301.65
}

```

