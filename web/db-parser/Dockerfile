# Build stage
FROM node:20-alpine AS build

WORKDIR /app

# Definir los argumentos que vendrán desde Dokploy
ARG VITE_BASE_PATH_URL
ARG VITE_ENDPOINT_URL

# Convertirlos en variables de entorno para que Vite los vea
ENV VITE_BASE_PATH_URL=$VITE_BASE_PATH_URL
ENV VITE_ENDPOINT_URL=$VITE_ENDPOINT_URL

# Copiar archivos de dependencias
COPY package*.json ./
RUN npm ci

# Copiar el resto del código y construir
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

# Diagnóstico: Ver archivos disponibles en el contexto actual
RUN ls -la

# Copiar configuración de Nginx primero
# Usamos el path relativo al contexto definido en Dokploy
COPY nginx.conf /etc/nginx/nginx.conf

# Copiar los archivos estáticos desde la etapa de build
COPY --from=build /app/dist /usr/share/nginx/html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]