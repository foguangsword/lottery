const hre = require("hardhat");

async function main() {
  const [deployer] = await hre.ethers.getSigners();
  console.log("Deploying with the Account:", deployer.address);
  const LuckyDraw = await hre.ethers.deployContract("MyLuckyDraw");
  console.log("Contract Address:", LuckyDraw.address);
}

main().catch((error) => {
  console.error(error);
  process.exitCode = 1;
});
