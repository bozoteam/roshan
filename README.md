# Projeto de Autenticação com JWT

Este projeto implementa um sistema de autenticação utilizando JSON Web Tokens (JWT) com as seguintes funcionalidades e práticas de segurança.

## Visão Geral

O projeto é construído em Go utilizando GORM para operações de banco de dados e Gin para lidar com requisições HTTP. Ele fornece endpoints para criação, atualização, listagem e exclusão de usuários, além de um sistema de autenticação JWT com suporte para refresh tokens.

## Funcionalidades

- **CRUD de Usuários**:
  - Criar, buscar, atualizar e deletar usuários.

- **Autenticação JWT**:
  - Geração de tokens JWT usando o algoritmo RS256 para segurança.
  - Tokens de acesso e refresh tokens para gerenciamento de sessão.

- **Proteção de Endpoints**:
  - Middleware para garantir que apenas usuários autenticados acessem certos endpoints.

- **Configuração via `.env`**:
  - Configurações sensíveis como chaves JWT são carregadas de um arquivo `.env`.

## Estrutura do Projeto

- **controllers/**: Contém a lógica de negócio e autenticação.
  - `userController.go`: Controladores para operações CRUD de usuários.
  - `authController.go`: Controlador para operações de autenticação e refresh de tokens.
  
- **helpers/**:
  - `tokenHelper.go`: Funções auxiliares para geração de tokens JWT.

- **middlewares/**:
  - `authMiddleware.go`: Middleware para proteção de rotas via JWT.

- **models/**:
  - `user.go`: Estrutura do modelo de usuário.

- **routes/**: Configura as rotas da aplicação.
  - `userRouter.go`: Rotas relacionadas a operações de usuário.
  - `authRouter.go`: Rotas relacionadas a autenticação.

- **database/**: Configuração e conexão com o banco de dados.
  - `adapter.go`: Gerencia a conexão com o banco de dados.

- **migrations/**: Scripts SQL para criação do esquema do banco de dados.

## Configuração do Ambiente

- **Chaves JWT**:
  - Armazenadas no `.env` sob as variáveis `JWT_SECRET` e `JWT_REFRESH_SECRET`.

- **Expirações de Token**:
  - `JWT_TOKEN_EXPIRATION`: Tempo de expiração dos tokens de acesso em minutos.
  - `JWT_REFRESH_TOKEN_EXPIRATION`: Tempo de expiração dos refresh tokens em horas.

## Instalação e Execução

1. Clone o repositório.
2. Instale as dependências:
   ```bash
   go mod tidy
   ```
3. Configure o arquivo `.env` com suas chaves e configurações.
4. Execute as migrações para criar as tabelas do banco de dados com Goose:
   ```bash
   goose up
   ```
5. Inicie o servidor:
   ```bash
   go run main.go
   ```