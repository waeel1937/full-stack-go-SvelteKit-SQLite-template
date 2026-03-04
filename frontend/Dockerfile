FROM node:20-alpine AS build
WORKDIR /app
COPY package.json ./
RUN npm install --legacy-peer-deps
COPY . .
RUN npm run build

FROM node:20-alpine
RUN addgroup -S edge && adduser -S edge -G edge
WORKDIR /app
COPY --from=build /app/build /app/build
COPY --from=build /app/package.json /app/
COPY --from=build /app/node_modules /app/node_modules
USER edge
EXPOSE 3000
ENV PORT=3000 HOST=0.0.0.0
CMD ["node", "build/index.js"]
