# Plan de Implementación: Conversor SQL a DynamoDB

## Resumen

Este plan de implementación convierte el diseño del conversor serverless SQL a DynamoDB en tareas de codificación discretas. El enfoque sigue una arquitectura serverless-first usando funciones Lambda en Go, Amazon Bedrock para conversión potenciada por IA, procesamiento asíncrono con SQS, y almacenamiento efímero en DynamoDB. Cada tarea construye incrementalmente hacia un sistema completo y listo para producción.

## Tareas

- [ ] 1. Configurar estructura del proyecto e interfaces principales
  - Crear módulo Go con estructura de directorios apropiada (cmd/, internal/, pkg/)
  - Definir interfaces principales para SQL_Parser, DynamoDB_Designer, Terraform_Generator
  - Configurar framework de testing con testify y gopter para pruebas basadas en propiedades
  - Crear tipos compartidos y modelos de datos para requests/responses
  - Definir estructura de entidad Conversion (id, status, input, output, timestamps, ttl)
  - _Requerimientos: 1.1, 2.1, 8.1, 8.2, 8.3, 9.1_

- [ ] 2. Implementar componente SQL_Parser
  - [ ] 2.1 Crear lógica de validación y parsing de SQL
    - Implementar parser de sentencias SQL CREATE TABLE de PostgreSQL
    - Agregar soporte para tipos de datos comunes, constraints y llaves foráneas
    - Crear representación estructurada de esquemas parseados
    - _Requerimientos: 1.2, 1.4_
  
  - [ ]* 2.2 Escribir prueba de propiedades para validación de sintaxis SQL
    - **Propiedad 1: Corrección de Validación de Sintaxis SQL**
    - **Valida: Requerimientos 1.2, 1.3**
  
  - [ ]* 2.3 Escribir pruebas unitarias para casos límite del parser SQL
    - Probar ejemplos específicos de sintaxis SQL y condiciones de error
    - Probar características no soportadas y validación de constraints
    - _Requerimientos: 1.2, 1.3_

- [ ] 3. Implementar componente DynamoDB_Designer
  - [ ] 3.1 Crear lógica principal de diseño de tablas DynamoDB
    - Implementar generación de estructura de tabla desde SQL parseado
    - Agregar soporte para selección de partition key y sort key
    - Crear manejo de patrones de optimización (read_heavy, write_heavy, balanced)
    - _Requerimientos: 2.2, 2.3, 2.4, 2.5_
  
  - [ ] 3.2 Implementar motor de recomendación de GSI
    - Analizar relaciones de llaves foráneas para oportunidades de GSI
    - Generar recomendaciones de partition key y sort key para GSIs
    - Crear análisis de patrones de acceso desde relaciones del esquema SQL
    - _Requerimientos: 3.1, 3.2, 4.2_
  
  - [ ]* 3.3 Escribir prueba de propiedades para influencia de patrón de optimización
    - **Propiedad 3: Influencia del Patrón de Optimización**
    - **Valida: Requerimientos 2.2, 2.3, 2.4, 2.5**
  
  - [ ]* 3.4 Escribir prueba de propiedades para generación de GSI
    - **Propiedad 5: Generación y Documentación de GSI**
    - **Valida: Requerimientos 3.1, 3.2, 3.3**
  
  - [ ] 3.5 Implementar lógica de patrón de diseño de tabla única
    - Analizar relaciones de entidades para oportunidades de tabla única
    - Crear estrategias de claves compuestas para entidades relacionadas
    - Agregar recomendaciones de desnormalización para relaciones complejas
    - _Requerimientos: 3.4, 4.5_
  
  - [ ]* 3.6 Escribir prueba de propiedades para aplicación de diseño de tabla única
    - **Propiedad 6: Aplicación de Diseño de Tabla Única**
    - **Valida: Requerimientos 3.4**

- [ ] 4. Implementar componente Terraform_Generator
  - [ ] 4.1 Crear generación de código Terraform para recursos DynamoDB
    - Generar bloques de recurso terraform para tablas DynamoDB
    - Incluir definiciones de GSI y configuración de modo de facturación
    - Agregar generación de políticas IAM para acceso a tablas
    - _Requerimientos: 5.1, 5.2, 5.3_
  
  - [ ] 4.2 Implementar estructura modular de Terraform y mejores prácticas
    - Crear módulos Terraform reutilizables
    - Agregar definiciones de variables y declaraciones de outputs
    - Incluir configuraciones de alarmas CloudWatch y backups
    - _Requerimientos: 5.4_
  
  - [ ]* 4.3 Escribir prueba de propiedades para validez de código Terraform
    - **Propiedad 11: Validez y Completitud del Código Terraform**
    - **Valida: Requerimientos 5.1, 5.2, 5.3**
  
  - [ ]* 4.4 Escribir pruebas unitarias para generación de Terraform
    - Probar ejemplos específicos de esquemas y casos límite
    - Probar configuraciones de optimización de facturación
    - _Requerimientos: 5.5_

- [ ] 5. Checkpoint - Validación de componentes principales
  - Asegurar que todas las pruebas pasen, preguntar al usuario si surgen dudas.

- [ ] 6. Implementar integración con Amazon Bedrock
  - [ ] 6.1 Crear cliente Bedrock e integración con API de Claude
    - Configurar AWS SDK para servicio Bedrock
    - Implementar invocación de modelo Claude con ingeniería de prompts apropiada
    - Agregar parsing y validación de respuestas para diseños generados por IA
    - _Requerimientos: 2.1, 2.6_
  
  - [ ] 6.2 Implementar análisis de esquemas potenciado por IA
    - Crear prompts para análisis de esquemas SQL y conversión a DynamoDB
    - Agregar inyección de contexto para patrones de optimización y requerimientos
    - Implementar validación de respuestas y manejo de errores para fallos de IA
    - _Requerimientos: 2.1, 2.2_
  
  - [ ]* 6.3 Escribir prueba de propiedades para procesamiento de SQL válido
    - **Propiedad 2: Aceptación de Procesamiento de SQL Válido**
    - **Valida: Requerimientos 1.4, 2.1**
  
  - [ ]* 6.4 Escribir pruebas unitarias para integración con Bedrock
    - Probar manejo de errores del servicio IA y lógica de reintentos
    - Probar configuraciones de timeout y mecanismos de fallback
    - _Requerimientos: 10.3, 10.4_

- [ ] 7. Implementar Conversion_Store (DynamoDB)
  - [ ] 7.1 Crear repositorio DynamoDB para conversiones
    - Implementar cliente DynamoDB con AWS SDK v2
    - Crear operaciones CRUD para entidad Conversion
    - Implementar actualización de estados (PENDING, PROCESSING, COMPLETED, FAILED)
    - _Requerimientos: 9.1, 9.2, 9.3, 9.4_
  
  - [ ] 7.2 Implementar consultas por fecha y TTL
    - Crear query por GSI de fecha (date_partition) para listar conversiones del día
    - Configurar TTL de 24 horas en todos los registros
    - Implementar GetItem por conversion_id para detalle
    - _Requerimientos: 9.5, 9.6, 9.7_
  
  - [ ]* 7.3 Escribir pruebas unitarias para repositorio DynamoDB
    - Probar operaciones CRUD con DynamoDB Local
    - Probar queries por fecha y manejo de TTL
    - _Requerimientos: 9.1, 9.5_

- [ ] 8. Implementar Conversion_Queue (SQS)
  - [ ] 8.1 Crear cliente SQS para encolar conversiones
    - Implementar SendMessage para nuevas conversiones
    - Configurar Dead Letter Queue para mensajes fallidos
    - Agregar atributos de mensaje (conversion_id, timestamp)
    - _Requerimientos: 6.4, 10.5_
  
  - [ ]* 8.2 Escribir pruebas unitarias para integración SQS
    - Probar encolado y estructura de mensajes
    - Probar configuración de DLQ
    - _Requerimientos: 6.4, 10.5_

- [ ] 9. Implementar función Lambda API_Handler
  - [ ] 9.1 Crear handler Lambda para endpoints de API
    - Implementar handler POST /convert (validar SQL, guardar PENDING, encolar SQS, retornar ID)
    - Implementar handler GET /conversions (query por fecha de hoy)
    - Implementar handler GET /conversions/{id} (obtener detalle completo)
    - Implementar handler GET /health (estado del sistema)
    - _Requerimientos: 8.1, 8.2, 8.3, 8.4, 8.5, 8.6_
  
  - [ ] 9.2 Implementar validación de origen y seguridad
    - Crear Origin_Verifier para validar header X-Origin-Verify
    - Agregar manejo de errores HTTP y códigos de estado apropiados
    - Implementar logging estructurado con correlation IDs
    - _Requerimientos: 7.1, 7.4, 11.1, 11.2_
  
  - [ ]* 9.3 Escribir prueba de propiedades para corrección de respuestas API
    - **Propiedad 17: Corrección de Respuestas API**
    - **Valida: Requerimientos 8.4, 8.5, 8.6**
  
  - [ ]* 9.4 Escribir prueba de propiedades para validación de seguridad
    - **Propiedad 15: Validación de Headers de Seguridad**
    - **Valida: Requerimientos 7.1, 7.4**

- [ ] 10. Implementar función Lambda Conversion_Worker
  - [ ] 10.1 Crear handler Lambda con trigger SQS
    - Implementar consumo de mensajes de Conversion_Queue
    - Actualizar estado a PROCESSING al iniciar
    - Integrar SQL_Parser, Conversion_Engine, DynamoDB_Designer, Terraform_Generator
    - _Requerimientos: 2.1, 6.2, 9.2_
  
  - [ ] 10.2 Implementar flujo completo de conversión
    - Parsear SQL de entrada
    - Invocar Bedrock Claude para análisis y conversión
    - Generar diseño DynamoDB con GSIs y patrones de acceso
    - Generar código Terraform
    - Actualizar Conversion_Store con resultado (COMPLETED o FAILED)
    - _Requerimientos: 2.6, 9.3, 9.4_
  
  - [ ] 10.3 Implementar manejo de errores y logging
    - Capturar errores de Bedrock y actualizar estado a FAILED
    - Registrar contexto detallado de errores incluyendo input SQL
    - Configurar timeout de 120 segundos para acomodar Bedrock
    - _Requerimientos: 10.3, 10.4, 11.3, 11.4_
  
  - [ ]* 10.4 Escribir prueba de propiedades para completitud de respuesta de conversión
    - **Propiedad 4: Completitud de Respuesta de Conversión**
    - **Valida: Requerimientos 2.6**

- [ ] 11. Checkpoint - Validación de funciones Lambda
  - Asegurar que todas las pruebas pasen, preguntar al usuario si surgen dudas.

- [ ] 12. Implementar integración con AWS Secrets Manager
  - [ ] 12.1 Crear gestión de secretos para configuración sensible
    - Implementar integración con cliente AWS Secrets Manager
    - Agregar recuperación de secreto de verificación de origen
    - Crear gestión de configuración para ambas funciones Lambda
    - _Requerimientos: 7.6_
  
  - [ ]* 12.2 Escribir prueba de propiedades para uso de gestión de secretos
    - **Propiedad 16: Uso de Gestión de Secretos**
    - **Valida: Requerimientos 7.6**

- [ ] 13. Implementar logging y monitoreo comprehensivo
  - [ ] 13.1 Crear logging estructurado con IDs de correlación
    - Implementar logging a CloudWatch para todas las operaciones API
    - Agregar generación y tracking de correlation ID (conversion_id)
    - Crear logging detallado de errores con información de contexto
    - _Requerimientos: 11.1, 11.2, 11.4_
  
  - [ ] 13.2 Implementar monitoreo de cola SQS
    - Monitorear profundidad de cola y edad de mensajes
    - Crear alertas para mensajes en DLQ
    - Implementar métricas de rendimiento de procesamiento
    - _Requerimientos: 11.5_
  
  - [ ]* 13.3 Escribir prueba de propiedades para logging comprehensivo
    - **Propiedad 20: Logging Comprehensivo**
    - **Valida: Requerimientos 11.1, 11.2, 11.4**
  
  - [ ]* 13.4 Escribir prueba de propiedades para disponibilidad de health check
    - **Propiedad 21: Disponibilidad de Health Check**
    - **Valida: Requerimientos 11.6**

- [ ] 14. Checkpoint - Validación de infraestructura de soporte
  - Asegurar que todas las pruebas pasen, preguntar al usuario si surgen dudas.

- [ ] 15. Crear aplicación web frontend
  - [ ] 15.1 Construir interfaz de aplicación de página única
    - Crear HTML/CSS/JavaScript para interfaz de entrada SQL con vista dividida
    - Implementar manejo de formularios para entrada de esquemas SQL (textarea y archivo)
    - Agregar UI de opciones de conversión (patrones de optimización, diseño de tabla única)
    - Implementar lista de "Conversiones del Día" con auto-refresh
    - _Requerimientos: 1.1, 1.5, 1.6_
  
  - [ ] 15.2 Implementar visualización de resultados de conversión
    - Crear UI para mostrar diseños de tablas DynamoDB
    - Agregar visualización de recomendaciones de GSI y patrones de acceso
    - Implementar display de código Terraform con resaltado de sintaxis
    - Mostrar estados de conversión (PENDING, PROCESSING, COMPLETED, FAILED)
    - _Requerimientos: 2.6, 3.3, 4.1, 9.7_
  
  - [ ] 15.3 Implementar polling y manejo de estados
    - Crear lógica de polling para conversiones en progreso
    - Implementar backoff exponencial para polling eficiente
    - Agregar indicadores de carga y estados de progreso
    - _Requerimientos: 9.7_
  
  - [ ] 15.4 Agregar manejo de errores y feedback al usuario
    - Implementar display de mensajes de error para fallos de validación
    - Crear mensajes de error amigables y sugerencias
    - _Requerimientos: 1.3, 8.5_

- [ ] 16. Crear código de infraestructura Terraform
  - [ ] 16.1 Crear infraestructura de funciones Lambda
    - Escribir Terraform para ambas funciones Lambda con configuración ARM64
    - Agregar integración con API Gateway y roles IAM
    - Incluir grupos de logs CloudWatch y configuración de monitoreo
    - Configurar timeouts (10s para API_Handler, 120s para Conversion_Worker)
    - _Requerimientos: 6.1, 6.2, 6.3, 10.3, 10.4_
  
  - [ ] 16.2 Crear infraestructura de SQS y DynamoDB
    - Escribir Terraform para Conversion_Queue con DLQ
    - Crear tabla Conversion_Store con GSI por fecha y TTL habilitado
    - Configurar políticas IAM para acceso entre servicios
    - _Requerimientos: 6.4, 6.5, 9.5, 9.6_
  
  - [ ] 16.3 Crear infraestructura de seguridad
    - Escribir Terraform para recursos de Secrets Manager
    - Incluir políticas IAM con acceso de mínimo privilegio
    - Agregar configuración de API Gateway con validación de origen
    - _Requerimientos: 7.5, 7.6_
  
  - [ ] 16.4 Agregar infraestructura de monitoreo y observabilidad
    - Crear dashboards y alarmas de CloudWatch
    - Agregar monitoreo de rendimiento y alertas
    - Incluir monitoreo de métricas de SQS (profundidad de cola, edad de mensajes)
    - _Requerimientos: 11.3, 11.5_

- [ ] 17. Implementar pipeline de build y deployment
  - [ ] 17.1 Crear proceso de build de Go para funciones Lambda
    - Configurar cross-compilation para arquitectura ARM64
    - Crear scripts de build para ambas funciones Lambda
    - Agregar gestión de dependencias y vendoring
    - _Requerimientos: 6.3_
  
  - [ ] 17.2 Crear proceso de build y deployment de frontend
    - Implementar optimización y minificación de assets frontend
    - Crear proceso de deployment a hosting estático (Cloudflare Pages o S3+CloudFront)
    - Agregar invalidación de caché y verificación de deployment
    - _Requerimientos: 1.1, 10.1_

- [ ] 18. Pruebas de integración y validación del sistema
  - [ ]* 18.1 Escribir pruebas de integración para flujos end-to-end
    - Probar flujo completo de conversión SQL a DynamoDB (submit → poll → result)
    - Probar enforcement de seguridad y manejo de errores
    - Probar integración de IA y manejo de timeouts
    - Probar lista de conversiones del día y TTL
    - _Requerimientos: Todos los requerimientos_
  
  - [ ]* 18.2 Escribir pruebas de propiedades para análisis de patrones de acceso
    - **Propiedad 8: Completitud de Documentación de Patrones de Acceso**
    - **Valida: Requerimientos 4.1, 4.3, 4.4**
  
  - [ ]* 18.3 Escribir pruebas de propiedades para manejo de relaciones complejas
    - **Propiedad 10: Manejo de Relaciones Complejas**
    - **Valida: Requerimientos 4.5**

- [ ] 19. Checkpoint final - Validación completa del sistema
  - Asegurar que todas las pruebas pasen, preguntar al usuario si surgen dudas.

## Notas

- Las tareas marcadas con `*` son opcionales y pueden omitirse para un MVP más rápido
- Cada tarea referencia requerimientos específicos para trazabilidad
- Los checkpoints aseguran validación incremental y feedback del usuario
- Las pruebas de propiedades validan propiedades de corrección universal con mínimo 100 iteraciones
- Las pruebas unitarias validan ejemplos específicos y casos límite
- El proceso de build apunta a ARM64 Graviton2 para optimización de costos
- Todos los componentes implementan manejo de errores comprehensivo y medidas de seguridad
- El modelo "Conversiones del Día" elimina la necesidad de autenticación y sesiones de usuario
- El TTL de 24 horas asegura limpieza automática sin intervención manual