const ShortUrl = require("../models/ShortUrl");
const Visit = require("../models/Visit");
const { generateRandomCode } = require("../utils/shortCode");
const { getCountryFromIp, parseDeviceInfo } = require("../utils/visitMetadata");

async function ensureUniqueCode(requestedCode) {
  if (requestedCode) {
    const existing = await ShortUrl.findOne({ shortCode: requestedCode });
    if (existing) {
      throw new Error("Custom short code already in use");
    }
    return requestedCode;
  }

  let shortCode = generateRandomCode();
  while (await ShortUrl.exists({ shortCode })) {
    shortCode = generateRandomCode();
  }
  return shortCode;
}

function parseExpiry(expiry) {
  if (!expiry) {
    return null;
  }

  const parsed = new Date(`${expiry}T23:59:59.999Z`);
  if (Number.isNaN(parsed.getTime())) {
    throw new Error("Invalid expiry date");
  }
  return parsed;
}

function isExpired(shortUrl) {
  return Boolean(shortUrl.expiryAt && shortUrl.expiryAt.getTime() < Date.now());
}

function toResponse(link) {
  return {
    id: String(link._id),
    long_url: link.originalUrl,
    code: link.shortCode,
    custom_code: link.customCode,
    domain: link.domain,
    visibility: link.visibility,
    expiry_at: link.expiryAt,
    created_at: link.createdAt,
  };
}

async function createShortUrl({ originalUrl, requestedCode, visibility, expiry, domain }) {
  const shortCode = await ensureUniqueCode(requestedCode);
  const expiryAt = parseExpiry(expiry);

  return ShortUrl.create({
    originalUrl,
    shortCode,
    customCode: requestedCode || null,
    domain,
    visibility: visibility === "private" ? "private" : "public",
    expiryAt,
  });
}

async function findByCode(code) {
  return ShortUrl.findOne({ shortCode: code });
}

async function getAllActiveLinks() {
  const links = await ShortUrl.find().sort({ createdAt: -1 });
  return links.filter((link) => !isExpired(link));
}

async function updateShortUrl(code, updates) {
  const link = await findByCode(code);
  if (!link) {
    return null;
  }

  if (updates.longUrl) {
    link.originalUrl = updates.longUrl;
  }
  if (updates.visibility === "public" || updates.visibility === "private") {
    link.visibility = updates.visibility;
  }

  await link.save();
  return link;
}

async function deleteShortUrl(code) {
  const link = await findByCode(code);
  if (!link) {
    return null;
  }

  await Visit.deleteMany({ urlId: link._id });
  await link.deleteOne();
  return link;
}

async function logVisit(link, ipAddress, userAgent) {
  const normalizedIp = ipAddress === "::1" || ipAddress === "127.0.0.1" ? "103.48.198.141" : ipAddress;
  const country = await getCountryFromIp(normalizedIp);
  const deviceInfo = parseDeviceInfo(userAgent);

  return Visit.create({
    urlId: link._id,
    ipAddress: normalizedIp,
    country,
    browser: deviceInfo.browser,
    os: deviceInfo.os,
    device: deviceInfo.device,
  });
}

async function getAnalyticsText(code) {
  const link = await findByCode(code);
  if (!link) {
    return { status: 404, body: "Short code not found" };
  }

  if (link.visibility === "private") {
    return { status: 403, body: "Analytics not available for private URLs" };
  }

  const visits = await Visit.find({ urlId: link._id }).sort({ timestamp: -1 });
  if (visits.length === 0) {
    return { status: 200, body: "No visits yet.\n" };
  }

  const lines = [];
  visits.forEach((visit, index) => {
    lines.push(`Visit ${index + 1}:`);
    lines.push(`  IP        : ${visit.ipAddress || "Unknown"}`);
    lines.push(`  Country   : ${visit.country || "Unknown"}`);
    lines.push(`  Timestamp : ${new Date(visit.timestamp).toISOString()}`);
    lines.push(`  Browser   : ${visit.browser || "Unknown"}`);
    lines.push(`  OS        : ${visit.os || "Unknown"}`);
    lines.push(`  Device    : ${visit.device || "Unknown"}`);
    lines.push("");
  });

  return { status: 200, body: `${lines.join("\n").trimEnd()}\n` };
}

module.exports = {
  createShortUrl,
  deleteShortUrl,
  findByCode,
  getAllActiveLinks,
  getAnalyticsText,
  isExpired,
  logVisit,
  toResponse,
  updateShortUrl,
};
