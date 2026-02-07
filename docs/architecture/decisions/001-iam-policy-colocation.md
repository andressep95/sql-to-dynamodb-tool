# ADR-001: Colocación de Políticas IAM junto a Recursos

**Estado:** Aceptado
**Fecha:** 2025-02-06
**Contexto:** Organización de políticas IAM en Terraform

---

## Contexto

Al diseñar la infraestructura con Terraform, surgió la pregunta de cómo organizar las políticas IAM para las funciones Lambda:

1. **Opción A - Módulo IAM centralizado:** Crear un módulo genérico `modules/iam/` que encapsule la creación de roles y políticas
2. **Opción B - Colocación junto a recursos:** Crear las políticas IAM en los mismos archivos donde se definen los recursos que protegen

---

## Decisión

**Adoptamos la Opción B: Colocación de políticas IAM junto a los recursos que protegen.**

Las políticas IAM se definen en:
- `shared/dynamodb.tf` → Políticas de acceso a DynamoDB por Lambda
- `shared/sqs.tf` → Políticas de acceso a SQS por Lambda
- `modules/bedrock/` → Políticas de acceso a Bedrock (caso especial, ver abajo)

Los roles base se crean directamente en `main.tf` del ambiente.

---

## Justificación

### 1. AWS Prescriptive Guidance recomienda no crear thin wrappers

> *"You shouldn't create modules that are thin wrappers around other single resource types. If you have trouble finding a name for your module that's different from the name of the main resource type inside it, your module probably isn't creating a new abstraction―it's adding unnecessary complexity."*
>
> — [AWS Prescriptive Guidance: Terraform Best Practices](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/structure.html)

Un módulo IAM genérico sería un thin wrapper sobre `aws_iam_role` y `aws_iam_role_policy`.

### 2. Encapsular relaciones lógicas, no recursos individuales

> *"Group sets of related resources such as networking foundations, data tiers, security controls, and applications."*

Las políticas de DynamoDB están lógicamente relacionadas con la tabla DynamoDB, no con otras políticas IAM.

### 3. Facilita auditoría de Least Privilege

Con colocación:
```
shared/dynamodb.tf
├── module "schemas_table"           ← El recurso
├── aws_iam_role_policy.process_handler_dynamodb    ← Solo PutItem
├── aws_iam_role_policy.conversion_worker_dynamodb  ← Solo UpdateItem
├── aws_iam_role_policy.query_handler_dynamodb      ← GetItem, Query, Scan
└── aws_iam_role_policy.dlq_handler_dynamodb        ← Solo UpdateItem
```

Es inmediatamente visible qué permisos tiene cada Lambda sobre DynamoDB.

### 4. Cambios atómicos

Si modificamos la tabla DynamoDB (ej: agregar un GSI), las políticas que necesitan actualizarse están en el mismo archivo.

---

## Excepciones

### Módulo Bedrock

El módulo `modules/bedrock/` SÍ incluye políticas IAM porque:
- Encapsula lógica compleja (ARNs de inference profiles cross-region)
- El patrón de ARN requiere transformaciones (`us.anthropic.` → `anthropic.`)
- Es reutilizable para múltiples modelos (Sonnet, Haiku)

```hcl
module "bedrock" {
  source           = "../../modules/bedrock"
  model_id         = "us.anthropic.claude-3-5-sonnet-20241022-v2:0"
  lambda_role_name = aws_iam_role.conversion_worker.name
  # La política se crea internamente con la lógica de ARN correcta
}
```

---

## Consecuencias

### Positivas
- Código más legible y auditable
- Least privilege más fácil de verificar
- Cambios relacionados en el mismo archivo
- Menos módulos que mantener

### Negativas
- Políticas repetidas si múltiples Lambdas necesitan permisos idénticos (mitigado: cada Lambda tiene permisos distintos)
- Sin abstracción reutilizable para patrones IAM (mitigado: no tenemos patrones repetidos)

---

## Referencias

- [AWS Prescriptive Guidance - Terraform Structure](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/structure.html)
- [AWS Prescriptive Guidance - Terraform Security](https://docs.aws.amazon.com/prescriptive-guidance/latest/terraform-aws-provider-best-practices/security.html)
- [AWS Lambda - Managing Permissions](https://docs.aws.amazon.com/lambda/latest/dg/lambda-permissions.html)
- [IAM Best Practices - Least Privilege](https://docs.aws.amazon.com/IAM/latest/UserGuide/best-practices.html#grant-least-privilege)
