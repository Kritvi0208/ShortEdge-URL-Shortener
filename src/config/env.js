const path = require("path");
const dotenv = require("dotenv");

dotenv.config({ path: path.resolve(process.cwd(), ".env") });

module.exports = {
  port: Number(process.env.PORT || 8080),
  mongoUri: process.env.MONGODB_URI || "mongodb://127.0.0.1:27017/shortedge",
};
