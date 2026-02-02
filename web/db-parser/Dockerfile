# Build stage
FROM node:20-alpine AS build

WORKDIR /app

# Argumentos de Dokploy
ARG VITE_BASE_PATH_URL
ARG VITE_ENDPOINT_URL
ENV VITE_BASE_PATH_URL=$VITE_BASE_PATH_URL
ENV VITE_ENDPOINT_URL=$VITE_ENDPOINT_URL

# 1. Copiamos los archivos de configuración desde la raíz del REPO
COPY web/db-parser/package*.json ./

# 2. Instalamos dependencias
RUN npm ci

# 3. Copiamos todo el contenido de la subcarpeta al WORKDIR (/app)
COPY web/db-parser/ .

# 4. Construimos (Asegúrate de haber quitado el type-check del package.json como hablamos antes)
RUN npm run build

# Production stage
FROM nginx:alpine

# 5. IMPORTANTE: En la etapa final, Docker vuelve a mirar la raíz del repo. 
# Como el contexto es la raíz, buscamos el nginx.conf en su ruta completa.
COPY web/db-parser/nginx.conf /etc/nginx/nginx.conf

# 6. Copiamos el resultado del build
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]