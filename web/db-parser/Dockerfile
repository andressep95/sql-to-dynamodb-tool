# Build stage
FROM node:20-alpine AS build

WORKDIR /app

# Definir los argumentos que vendrán desde Dokploy
ARG VITE_BASE_PATH_URL
ARG VITE_ENDPOINT_URL

# Convertirlos en variables de entorno para que Vite los vea
ENV VITE_BASE_PATH_URL=$VITE_BASE_PATH_URL
ENV VITE_ENDPOINT_URL=$VITE_ENDPOINT_URL

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine

COPY --from=build /app/dist /usr/share/nginx/html
# IMPORTANTE: Asegúrate de que el archivo se llame nginx.conf en tu carpeta
COPY ./nginx.conf /etc/nginx/nginx.conf
EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]