const { app, connectDatabase } = require("../src/app");

let bootPromise;

module.exports = async (req, res) => {
  try {
    if (!bootPromise) {
      bootPromise = connectDatabase();
    }

    await bootPromise;
    return app(req, res);
  } catch (error) {
    console.error("Failed to initialize Vercel handler:", error);
    return res.status(500).json({ error: "Failed to initialize server" });
  }
};
