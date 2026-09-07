# Разработка: Nuxt с горячей перезагрузкой.
# pnpm, а не npm: на графе зависимостей Nuxt разрешатель npm 10 падает
# с «Cannot read properties of null (reading 'edgesOut')».
FROM node:22-alpine

RUN npm install -g pnpm@10

WORKDIR /app

# Слой зависимостей кэшируется отдельно от исходников
COPY package.json ./
RUN pnpm install

COPY . .

ENV NODE_ENV=development \
    NUXT_HOST=0.0.0.0 \
    NUXT_PORT=3000 \
    MOCKS_DIR=/mocks

EXPOSE 3000
CMD ["pnpm", "run", "dev"]

# Продакшен (когда дойдёт до деплоя):
#   RUN pnpm install --frozen-lockfile && pnpm run build
#   CMD ["node", ".output/server/index.mjs"]
# Сборка съедает 1–2 ГБ, поэтому собирать надо в CI,
# а не на шестигигабайтном сервере рядом с постгресом.
