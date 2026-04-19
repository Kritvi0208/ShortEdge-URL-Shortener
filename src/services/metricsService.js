const client = require("prom-client");

client.collectDefaultMetrics();

const shortenedCounter = new client.Counter({
  name: "url_shortened_total",
  help: "Total number of short links created",
});

const redirectCounter = new client.Counter({
  name: "url_redirect_total",
  help: "Total number of redirects",
});

module.exports = {
  register: client.register,
  shortenedCounter,
  redirectCounter,
};
