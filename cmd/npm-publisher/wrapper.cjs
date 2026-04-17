#!/usr/bin/env node

// This script is the entry point for the @ghost.build/cli npm package.
// It resolves and spawns the correct platform-specific binary from
// the @ghost.build/cli-{os}-{arch} optional dependency packages.

const { platform, arch, env, argv } = process;
const { spawnSync } = require("node:child_process");

const PLATFORMS = {
  darwin: {
    x64: "@ghost.build/cli-darwin-x64/bin/ghost",
    arm64: "@ghost.build/cli-darwin-arm64/bin/ghost",
  },
  linux: {
    x64: "@ghost.build/cli-linux-x64/bin/ghost",
    arm64: "@ghost.build/cli-linux-arm64/bin/ghost",
  },
  win32: {
    x64: "@ghost.build/cli-win32-x64/bin/ghost.exe",
  },
};

const binPath = env.GHOST_BINARY || PLATFORMS[platform]?.[arch];

if (!binPath) {
  console.error(
    `ghost does not ship prebuilt binaries for your platform (${platform}-${arch}).`
  );
  console.error(
    "For more information, visit: https://ghost.build/docs"
  );
  process.exitCode = 1;
} else {
  let resolvedPath;
  try {
    resolvedPath = env.GHOST_BINARY || require.resolve(binPath);
  } catch {
    console.error(
      `Could not find the ghost binary for your platform (${platform}-${arch}).`
    );
    console.error(
      "The platform-specific package may not have been installed correctly."
    );
    console.error("Try reinstalling: npm install -g @ghost.build/cli");
    process.exitCode = 1;
  }

  if (resolvedPath) {
    const result = spawnSync(resolvedPath, argv.slice(2), {
      shell: false,
      stdio: "inherit",
    });

    if (result.error) {
      throw result.error;
    }

    process.exitCode = result.status;
  }
}
