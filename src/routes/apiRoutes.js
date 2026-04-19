const express = require("express");
const mongoose = require("mongoose");
const {
  createShortUrl,
  deleteShortUrl,
  findByCode,
  getAllActiveLinks,
  getAnalyticsText,
  isExpired,
  logVisit,
  toResponse,
  updateShortUrl,
} = require("../services/shortUrlService");
const { register, redirectCounter, shortenedCounter } = require("../services/metricsService");
const { getBaseUrl, getClientIp } = require("../utils/request");

const router = express.Router();

router.get("/health", async (req, res) => {
  if (mongoose.connection.readyState !== 1) {
    return res.status(500).json({ status: "db unreachable" });
  }

  return res.json({ status: "ok" });
});

router.get("/metrics", async (req, res) => {
  res.set("Content-Type", register.contentType);
  res.send(await register.metrics());
});

router.post("/shorten", async (req, res) => {
  try {
    const originalUrl = (req.body.url || "").trim();
    const requestedCode = (req.body.code || "").trim();
    const visibility = (req.body.visibility || "public").trim().toLowerCase();
    const expiry = (req.body.expiry || "").trim();

    if (!originalUrl) {
      return res.status(400).send("URL cannot be empty");
    }

    const link = await createShortUrl({
      originalUrl,
      requestedCode,
      visibility,
      expiry,
      domain: req.get("host"),
    });

    shortenedCounter.inc();
    const baseUrl = getBaseUrl(req);

    return res.json({
      short_url: `${baseUrl}/r/${link.shortCode}`,
      analytics_url: `${baseUrl}/analytics/${link.shortCode}`,
    });
  } catch (error) {
    const status = error.message === "Custom short code already in use" || error.message === "Invalid expiry date" ? 400 : 500;
    return res.status(status).send(error.message || "Failed to create short link");
  }
});

router.get("/r/:code", async (req, res) => {
  const link = await findByCode(req.params.code);
  if (!link) {
    return res.status(404).send("URL not found");
  }

  if (isExpired(link)) {
    return res.status(410).send("This short link has expired");
  }

  await logVisit(link, getClientIp(req), req.get("user-agent"));
  redirectCounter.inc();
  return res.redirect(link.originalUrl);
});

router.get("/analytics/:code", async (req, res) => {
  const result = await getAnalyticsText(req.params.code);
  return res.status(result.status).type("text/plain").send(result.body);
});

router.get("/all", async (req, res) => {
  const links = await getAllActiveLinks();
  return res.json(links.map(toResponse));
});

router.put("/update/:code", async (req, res) => {
  const updated = await updateShortUrl(req.params.code, {
    longUrl: (req.body.long_url || "").trim(),
    visibility: (req.body.visibility || "").trim().toLowerCase(),
  });

  if (!updated) {
    return res.status(404).send("Short code not found");
  }

  return res.type("text/plain").send(`Updated short link '${req.params.code}'`);
});

router.delete("/delete/:code", async (req, res) => {
  const deleted = await deleteShortUrl(req.params.code);
  if (!deleted) {
    return res.status(404).send("Short code not found");
  }

  return res.type("text/plain").send(`Deleted short link '${req.params.code}'`);
});

module.exports = router;
