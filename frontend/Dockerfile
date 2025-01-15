FROM node:20.11.0-alpine

RUN apk add --no-cache curl

WORKDIR /var/app

# Trick to make use of caching
COPY package.json package-lock.json /var/app/
RUN npm install

COPY . /var/app

CMD ["npm", "run", "dev"]
