FROM node:22-alpine

RUN apk add --no-cache bash openjdk17-jre \
  && npm install -g firebase-tools@15.22.4 \
  && npm cache clean --force

WORKDIR /workspace
COPY firebase.json .firebaserc ./

EXPOSE 9099 4000

CMD ["firebase", "emulators:start", "--only", "auth", "--project", "cineforge-local", "--config", "firebase.json"]
