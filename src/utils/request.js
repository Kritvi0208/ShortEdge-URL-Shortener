function getBaseUrl(req) {
  const protocol = req.headers["x-forwarded-proto"] || req.protocol;
  const host = req.headers["x-forwarded-host"] || req.get("host");
  return `${protocol}://${host}`;
}

function getClientIp(req) {
  const forwardedFor = req.headers["x-forwarded-for"];
  if (typeof forwardedFor === "string" && forwardedFor.length > 0) {
    return forwardedFor.split(",")[0].trim();
  }

  const realIp = req.headers["x-real-ip"];
  if (typeof realIp === "string" && realIp.length > 0) {
    return realIp.trim();
  }

  const ip = req.ip || req.socket?.remoteAddress || "127.0.0.1";
  return typeof ip === "string" && ip.startsWith("::ffff:") ? ip.slice(7) : ip;
}

module.exports = { getBaseUrl, getClientIp };
