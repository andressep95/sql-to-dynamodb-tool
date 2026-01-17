# SQL to DynamoDB Converter

Herramienta serverless que convierte esquemas SQL relacionales a diseños optimizados de DynamoDB usando IA, generando automáticamente tabla designs, índices secundarios globales (GSI) y código Terraform.

## El Problema

Migrar de bases de datos relacionales (SQL) a DynamoDB requiere:
- Rediseñar esquemas de relacional a NoSQL
- Identificar patrones de acceso óptimos
- Crear índices secundarios apropiados
- Configurar infraestructura como código

Este proceso es complejo, propenso a errores y requiere expertise en modelado DynamoDB.

## La Solución

Aplicación web que analiza tus `CREATE TABLE` statements SQL y genera:

- **DynamoDB Table Designs**: Esquemas optimizados con partition/sort keys
- **Global Secondary Indexes (GSI)**: Recomendaciones basadas en relaciones SQL
- **Access Patterns**: Documentación de cómo consultar eficientemente
- **Terraform Code**: IaC listo para desplegar en AWS
- **Análisis con IA**: Usa Amazon Bedrock (Claude) para optimizaciones inteligentes

## Stack Tecnológico

- **Backend**: AWS Lambda (Go) con ARM64 Graviton2
- **IA/ML**: Amazon Bedrock (Claude)
- **Frontend**: SPA estática servida desde Lambda + S3
- **Infraestructura**: Terraform modular (soporta LocalStack para dev, AWS para prod)
- **API Gateway**: REST v1 (LocalStack) / HTTP v2 (AWS)

## Arquitectura

```
Usuario → API Gateway → Lambda Converter → Bedrock (Claude)
                                          ↓
                               DynamoDB Design + Terraform
```

**Componentes principales:**
- **API Converter Lambda**: Procesa SQL, invoca IA, genera Terraform
- **Frontend Proxy Lambda**: Sirve interfaz web desde S3 privado
- **Bedrock Integration**: Claude analiza esquemas y sugiere optimizaciones

## Estructura del Proyecto

```
├── lambda/           # Funciones Go (converter, frontend-proxy)
├── web/              # Frontend SPA
├── infra/terraform/  # Módulos IaC
│   ├── modules/      # Lambda, API Gateway, IAM, S3, Bedrock
│   └── environments/ # Configuraciones dev/prod
└── spec/             # Requirements y plan de implementación
```

## Features

- **Validación SQL**: Detecta errores de sintaxis antes de convertir
- **Optimización configurable**: Read-heavy, write-heavy, balanced
- **Single-table design**: Sugiere patrones de tabla única cuando aplica
- **Seguridad**: Zero-trust, HTTPS end-to-end, IAM least privilege
- **Observabilidad**: CloudWatch logs, métricas, alarmas automáticas
- **Multi-ambiente**: LocalStack (dev) y AWS (producción)

## Casos de Uso

- Migración de aplicaciones legacy SQL a serverless DynamoDB
- Aprendizaje de modelado NoSQL desde esquemas relacionales conocidos
- Generación rápida de prototipos de infraestructura DynamoDB
- Análisis de patrones de acceso para schemas existentes

---

**Desarrollado con**: AWS Lambda, Amazon Bedrock, Terraform, Go
