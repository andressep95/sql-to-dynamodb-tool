# Requerimientos del Documento

## Introducción

El Conversor de SQL a DynamoDB es una aplicación web sin servidor que transforma esquemas relacionales de PostgreSQL en modelos de datos optimizados de DynamoDB mediante capacidades de IA/ML. El sistema proporciona un servicio de conversión inteligente que analiza las sentencias SQL CREATE TABLE y genera diseños completos de tablas de DynamoDB, incluyendo patrones de acceso, índices secundarios globales y código de infraestructura de Terraform.

La aplicación implementa un modelo de "Conversiones del Día" donde todas las conversiones son públicas, efímeras (TTL de 24 horas) y accesibles sin autenticación. El procesamiento es asíncrono mediante colas para manejar los tiempos de respuesta de la IA sin afectar la experiencia del usuario.

## Glosario

- **SQL_Parser**: Componente que valida y analiza las sentencias SQL de PostgreSQL antes del procesamiento
- **Conversion_Engine**: Servicio de IA/ML que utiliza Amazon Bedrock con Claude para la transformación de esquemas
- **API_Handler**: Función Lambda que gestiona las solicitudes HTTP (crear, listar, obtener conversiones) y la validación de origen
- **Conversion_Worker**: Función Lambda que procesa conversiones de forma asíncrona consumiendo mensajes de la cola SQS
- **DynamoDB_Designer**: Componente lógico dentro del Conversion_Engine que genera estructuras de tablas optimizadas de DynamoDB
- **Schemas_Table**: Tabla DynamoDB que almacena las conversiones del día con TTL automático de 24 horas
- **Conversion_Queue**: Cola SQS que desacopla la recepción de solicitudes del procesamiento con Bedrock

## Requirements

### Requirement 1: Entrada y Validación de Esquemas SQL

**User Story:** Como desarrollador, quiero ingresar sentencias SQL CREATE TABLE a través de una interfaz web, para poder convertir mis esquemas de base de datos relacional al formato DynamoDB.

#### Acceptance Criteria

1. WHEN un usuario accede a la aplicación web, THE sistema SHALL servir una interfaz de aplicación de página única con vista dividida (entrada/salida)
2. WHEN un usuario ingresa sentencias SQL CREATE TABLE, THE SQL_Parser SHALL validar la sintaxis antes del procesamiento
3. WHEN se proporciona sintaxis SQL inválida, THE SQL_Parser SHALL retornar mensajes de error descriptivos con números de línea
4. WHEN se envía SQL válido, THE API_Handler SHALL aceptar la entrada, almacenarla como PENDING y encolarla para procesamiento
5. WHEN un usuario sube un archivo .sql, THE API_Handler SHALL extraer el contenido de texto y procesarlo igual que el texto directo
6. THE sistema SHALL soportar entrada mediante texto plano en textarea o carga de archivos .sql

### Requirement 2: Conversión de Esquemas Potenciada por IA

**User Story:** Como arquitecto de bases de datos, quiero convertir esquemas SQL a diseños DynamoDB optimizados usando análisis de IA, para poder aprovechar las mejores prácticas de NoSQL automáticamente.

#### Acceptance Criteria

1. WHEN se recibe SQL válido, THE Conversion_Worker SHALL consumir el mensaje de SQS y analizar el esquema usando Amazon Bedrock con Claude
2. WHEN se analizan esquemas, THE DynamoDB_Designer SHALL generar estructuras de tablas optimizadas basadas en patrones de optimización especificados
3. WHERE la optimización está configurada como "read_heavy", THE DynamoDB_Designer SHALL priorizar el rendimiento de lectura en el diseño de tablas
4. WHERE la optimización está configurada como "write_heavy", THE DynamoDB_Designer SHALL priorizar el rendimiento de escritura en el diseño de tablas
5. WHERE la optimización está configurada como "balanced", THE DynamoDB_Designer SHALL balancear el rendimiento de lectura y escritura
6. WHEN la conversión está completa, THE Conversion_Worker SHALL actualizar el Conversion_Store con estado COMPLETED y el resultado JSON

### Requirement 3: Recomendaciones de Índices Secundarios Globales

**User Story:** Como desarrollador NoSQL, quiero recomendaciones automáticas de GSI basadas en mi esquema SQL, para poder soportar patrones de consulta eficientes en DynamoDB.

#### Acceptance Criteria

1. WHEN se analizan llaves foráneas e índices SQL, THE DynamoDB_Designer SHALL recomendar Índices Secundarios Globales apropiados
2. WHEN se generan GSIs, THE DynamoDB_Designer SHALL incluir recomendaciones de partition key y sort key
3. WHEN se crean recomendaciones de GSI, THE DynamoDB_Designer SHALL proporcionar explicaciones de patrones de acceso
4. THE DynamoDB_Designer SHALL soportar patrones de diseño de tabla única cuando sea beneficioso
5. WHEN existen múltiples patrones de consulta, THE DynamoDB_Designer SHALL optimizar el diseño de GSI para los patrones más comunes

### Requirement 4: Análisis de Patrones de Acceso

**User Story:** Como arquitecto de sistemas, quiero análisis detallado de patrones de acceso para mi esquema convertido, para poder entender cómo consultar mis tablas DynamoDB eficientemente.

#### Acceptance Criteria

1. WHEN la conversión está completa, THE DynamoDB_Designer SHALL generar documentación completa de patrones de acceso
2. WHEN se analizan relaciones, THE DynamoDB_Designer SHALL identificar patrones de acceso primarios desde el esquema SQL
3. WHEN se documentan patrones, THE DynamoDB_Designer SHALL proporcionar consultas de ejemplo para cada patrón de acceso
4. THE DynamoDB_Designer SHALL incluir consideraciones de rendimiento para cada patrón recomendado
5. WHEN existen relaciones complejas, THE DynamoDB_Designer SHALL sugerir estrategias de desnormalización

### Requirement 5: Generación de Código de Infraestructura

**User Story:** Como ingeniero DevOps, quiero código Terraform generado para mis recursos DynamoDB, para poder desplegar la infraestructura usando prácticas de Infraestructura como Código.

#### Acceptance Criteria

1. WHEN el diseño DynamoDB está completo, THE Terraform_Generator SHALL producir archivos de configuración Terraform válidos
2. WHEN se genera Terraform, THE Terraform_Generator SHALL incluir todas las tablas DynamoDB y GSIs recomendados
3. WHEN se crea código de infraestructura, THE Terraform_Generator SHALL incluir políticas IAM apropiadas para acceso a tablas
4. THE Terraform_Generator SHALL generar código Terraform modular siguiendo mejores prácticas
5. WHEN la optimización de facturación está habilitada, THE Terraform_Generator SHALL configurar ajustes de capacidad provisionada apropiados

### Requirement 6: Implementación de Arquitectura Sin Servidor

**User Story:** Como ingeniero de plataforma, quiero una arquitectura sin servidor que escale automáticamente y minimice la sobrecarga operacional, para que la aplicación pueda manejar cargas de trabajo variables de manera rentable.

#### Acceptance Criteria

1. THE API_Handler SHALL ejecutarse en AWS Lambda integrada con API Gateway para procesar solicitudes HTTP
2. THE Conversion_Worker SHALL ejecutarse en AWS Lambda con trigger de SQS para procesamiento asíncrono de conversiones
3. WHEN las funciones Lambda son desplegadas, THE sistema SHALL usar procesadores ARM64 Graviton2 para optimización de costos
4. THE Conversion_Queue SHALL desacoplar la recepción de solicitudes del procesamiento de Bedrock para evitar timeouts
5. THE Conversion_Store SHALL almacenar todas las conversiones con TTL de 24 horas para limpieza automática
6. WHEN las funciones son invocadas, THE sistema SHALL usar modelo de facturación por uso sin costos en reposo

### Requirement 7: Implementación de Seguridad y Zero Trust

**User Story:** Como ingeniero de seguridad, quiero controles de seguridad completos incluyendo principios Zero Trust, para que la aplicación esté protegida contra varios vectores de ataque.

#### Acceptance Criteria

1. THE Origin_Verifier SHALL validar todas las solicitudes usando encabezados personalizados X-Origin-Verify inyectados por Cloudflare
2. WHEN se sirve contenido, THE sistema SHALL aplicar cifrado HTTPS de extremo a extremo
3. THE Cloudflare_WAF SHALL proporcionar protección de Firewall de Aplicaciones Web y mitigación de DDoS
4. WHEN se accede a API Gateway, THE sistema SHALL rechazar solicitudes que no contengan el header X-Origin-Verify válido
5. THE sistema SHALL implementar roles IAM de mínimo privilegio para todos los componentes AWS
6. WHEN se almacena configuración sensible, THE sistema SHALL usar AWS Secrets Manager para el header de verificación

### Requirement 8: Implementación de Endpoints API

**User Story:** Como desarrollador frontend, quiero endpoints API bien definidos para toda la funcionalidad de la aplicación, para poder construir una interfaz de usuario responsiva.

#### Acceptance Criteria

1. THE API_Handler SHALL exponer un endpoint POST /convert para enviar esquemas SQL y recibir un ID de conversión
2. THE API_Handler SHALL exponer un endpoint GET /conversions para listar todas las conversiones del día actual
3. THE API_Handler SHALL exponer un endpoint GET /conversions/{id} para obtener el detalle completo de una conversión
4. THE API_Handler SHALL exponer un endpoint GET /health para monitoreo de disponibilidad del sistema
5. WHEN se llaman endpoints API, THE sistema SHALL retornar códigos de estado HTTP apropiados y mensajes de error
6. WHEN se procesa POST /convert, THE API_Handler SHALL retornar inmediatamente con {id, status: "PENDING"} sin esperar a Bedrock

### Requirement 9: Procesamiento Asíncrono y Modelo de Conversiones del Día

**User Story:** Como usuario final, quiero ver todas las conversiones del día en una lista pública, para poder reutilizar conversiones existentes y monitorear el progreso de mis solicitudes.

#### Acceptance Criteria

1. WHEN se envía una conversión, THE API_Handler SHALL crear un registro en Conversion_Store con estado PENDING y encolar en SQS
2. WHEN el Conversion_Worker procesa un mensaje, THE sistema SHALL actualizar el estado a PROCESSING durante el llamado a Bedrock
3. WHEN la conversión termina exitosamente, THE Conversion_Worker SHALL actualizar el estado a COMPLETED con el resultado
4. WHEN la conversión falla, THE Conversion_Worker SHALL actualizar el estado a FAILED con el mensaje de error
5. THE Conversion_Store SHALL usar un GSI con partition key fecha (YYYY-MM-DD) para consultas eficientes de conversiones del día
6. THE sistema SHALL configurar TTL de 24 horas en todos los registros para limpieza automática sin intervención manual
7. WHEN el cliente consulta GET /conversions/{id}, THE sistema SHALL retornar el estado actual y resultado si está disponible

### Requirement 10: Optimización de Rendimiento y Caché

**User Story:** Como usuario final, quiero tiempos de respuesta rápidos y uso eficiente de recursos, para poder convertir esquemas rápidamente sin demoras.

#### Acceptance Criteria

1. THE Cloudflare_WAF SHALL implementar estrategias de caché inteligente para activos estáticos del frontend
2. WHEN las funciones Lambda son invocadas, THE sistema SHALL optimizar el rendimiento de arranque en frío mediante binarios Go compilados
3. THE API_Handler SHALL tener timeout de 10 segundos para respuestas rápidas al cliente
4. THE Conversion_Worker SHALL tener timeout de 120 segundos para acomodar tiempos de respuesta de Bedrock
5. THE Conversion_Queue SHALL configurar Dead Letter Queue para reintentos de mensajes fallidos
6. WHEN se sirven solicitudes repetidas de lista, THE sistema SHALL aprovechar caché de borde de Cloudflare

### Requirement 11: Monitoreo y Observabilidad

**User Story:** Como administrador del sistema, quiero monitoreo y registro completos, para poder solucionar problemas y monitorear el rendimiento del sistema.

#### Acceptance Criteria

1. THE sistema SHALL registrar todas las solicitudes y respuestas API en AWS CloudWatch Logs
2. WHEN ocurren errores, THE sistema SHALL capturar información detallada de errores con IDs de correlación (conversion_id)
3. THE sistema SHALL monitorear métricas de rendimiento de funciones Lambda incluyendo duración, memoria y errores
4. WHEN la conversión de IA falla, THE sistema SHALL registrar contexto detallado del error incluyendo el input SQL para debugging
5. THE sistema SHALL monitorear profundidad de cola SQS y edad de mensajes para detectar problemas de procesamiento
6. THE sistema SHALL proporcionar endpoint /health para monitoreo de disponibilidad del sistema