require("./config/env");

const express = require("express");
const path = require("path");
const { connectDatabase } = require("./config/db");
const apiRoutes = require("./routes/apiRoutes");

const app = express();
const publicDir = path.join(process.cwd(), "public");

app.use(express.urlencoded({ extended: false }));
app.use(express.json());
app.use(express.static(publicDir));

app.get("/", (req, res) => {
  res.sendFile(path.join(publicDir, "index.html"));
});

app.use("/", apiRoutes);

app.get("*", (req, res) => {
  const requestPath = req.path.startsWith("/") ? req.path.slice(1) : req.path;
  const targetPath = requestPath || "index.html";
  res.sendFile(path.join(publicDir, targetPath), (error) => {
    if (error) {
      res.status(404).send("Page not found");
    }
  });
});

app.use((error, req, res, next) => {
  if (error instanceof SyntaxError && error.status === 400 && "body" in error) {
    return res.status(400).send("Invalid JSON body");
  }

  return next(error);
});

module.exports = { app, connectDatabase };
