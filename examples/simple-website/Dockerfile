FROM node:22-alpine AS build

WORKDIR /app

RUN corepack enable

COPY package.json pnpm-lock.yaml ./
RUN pnpm config set dangerouslyAllowAllBuilds true \
	&& pnpm install --frozen-lockfile

COPY . .

ARG SITE_URL=https://blog.domi.ninja
ENV SITE_URL=${SITE_URL}

RUN pnpm build

FROM nginx:1.27-alpine

COPY --from=build /app/dist/ /usr/share/nginx/html/

RUN cat > /etc/nginx/conf.d/default.conf <<'NGINX'
server {
    listen 80;
    server_name _;
    root /usr/share/nginx/html;
    index index.html;

    location / {
        try_files $uri $uri/ $uri.html =404;
    }
}
NGINX
