const { app, connectDatabase } = require("./src/app");
const { port } = require("./src/config/env");

connectDatabase()
  .then(() => {
    app.listen(port, () => {
      console.log(`ShortEdge server running on port ${port}`);
    });
  })
  .catch((error) => {
    console.error("Failed to start ShortEdge:", error.message);
    process.exit(1);
  });
