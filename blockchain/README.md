## MyLuckyDraw合约



#### 安装项目依赖

```bash
cd .\lottery\blockchain\
npm install
```

#### 编译与测试

```bash
npx hardhat compile
npx hardhat run .\scripts\run.js
```

会有类似如下输出，则说明测试脚本运行成功：

```
已开奖, 交易Hash: 0x3887c4d8936c263d825b5823ccb9ca4f142976eb75cdaa573551003801bbe88e， 区块Hash: 0x13e5a56a85ade94a925739006b3902ca1a66dfcc6bc71535dd9c88ae3b98050a，区块高度: 278
中奖人索引：[
  BigNumber { value: "52" },
  BigNumber { value: "155" },
  BigNumber { value: "19" },
  BigNumber { value: "63" },
  BigNumber { value: "189" }
]
中奖人地址：[
  '0xDe78a7f723C8159a7b81Dc946c6B70838805F667',
  '0x61922B9e1d5F789D4235b29049404E8691d2032c',
  '0xb5Bb5aF90c6E7Ba3c1Fa7b423554C0a5d9fea498',
  '0xA26e88e95a8F779c87Ff5aae8052d96d62750a91',
  '0xfF837678ddAFb639CfD5D3620A03126ce8441419'
]
```



#### 部署到ganache

安装好Ganache后启动，然后在blockchain目录下新建文件`.env`

```
GANACHE_SK0=GANACHE上的私钥
```

运行部署脚本：

```
npx hardhat run .\scripts\deploy.js --network ganache
```

输出类似:

```
Deploying with the Account: 0x788230E1790DB5d52728851adc46461E9ABa7424
Contract Address: 0xE7D6C332e7b78e19Bc36bD06596a36f332860f74
```

说明合约部署成功。

注意：部署到正式环境的时候建议把合约中的类似如下与hardhat console的代码去掉：

```solidity
import "hardhat/console.sol";
console.log("draw seed is:" , seed);  
console.log("luckIndex is: ", luckyIndex); 
console.log("winner is: ");                 
for(...) { console.log(...); } 
```

这些在 Hardhat 本地能跑，但部署到 测试网/主网 时：
  - 如果编译环境没有 Hardhat console 库，编译失败
  - 增加合约体积，浪费 gas