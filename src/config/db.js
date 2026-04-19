const mongoose = require("mongoose");
const { mongoUri } = require("./env");

let connectionPromise;

function connectDatabase() {
  if (!connectionPromise) {
    connectionPromise = mongoose.connect(mongoUri, {
      serverSelectionTimeoutMS: 5000,
    });
  }

  return connectionPromise;
}

module.exports = { connectDatabase };
