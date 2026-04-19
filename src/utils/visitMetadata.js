const UAParser = require("ua-parser-js");

async function getCountryFromIp(ipAddress) {
  try {
    const response = await fetch(`https://ipwho.is/${encodeURIComponent(ipAddress)}`);
    if (!response.ok) {
      return "Unknown";
    }

    const payload = await response.json();
    return payload.success ? payload.country || "Unknown" : "Unknown";
  } catch (error) {
    return "Unknown";
  }
}

function parseDeviceInfo(userAgent) {
  const parser = new UAParser(userAgent || "");
  const result = parser.getResult();

  let device = "PC";
  if (result.device.type === "mobile") {
    device = "Mobile";
  } else if (result.device.type === "tablet") {
    device = "Tablet";
  }

  return {
    browser: result.browser.name || "Unknown",
    os: result.os.name || "Unknown",
    device,
  };
}

module.exports = { getCountryFromIp, parseDeviceInfo };
