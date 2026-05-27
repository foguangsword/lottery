const main = async () => {
    const signers = await hre.ethers.getSigners();
    console.log(signers.length);
    const owner = signers[0];
    const contractFactory = await hre.ethers.getContractFactory('MyLuckyDraw');
    const contract = await contractFactory.deploy();
    await contract.deployed();
    console.log("Contract deployed to: %s, owner: %s", contract.address, owner.address);

    //创建活动
    const startTime = parseInt(new Date().getTime()/1000);
    const endTime = parseInt(new Date().getTime()/1000 + 3600);
    let tx = await contract.createActivity("activity1", 
                    startTime, 
                    endTime, 
                    5);
    let receipt = await tx.wait();
    if(receipt.status === 1)
      console.log("活动创建成功: %s", await contract.allDraws("activity1"));
    else
      console.log("活动创建失败: %s", receipt);

    //把时间调到 startTime + 1，确保报名窗口已打开
    await hre.network.provider.send("evm_setNextBlockTimestamp", [startTime + 5]);
    await hre.network.provider.send("evm_mine");
    
    //报名
    for(var i=0; i<273; i++){
      const wallet = await hre.ethers.Wallet.createRandom();
      tx = await contract.participate("activity1", wallet.address);
      await tx.wait();
      //console.log("报名人", wallet.address);
    }

    // 把时间调到 endTime + 1，确保已过截止期
    await hre.network.provider.send("evm_setNextBlockTimestamp", [endTime + 5]);
    await hre.network.provider.send("evm_mine");
    
    //开奖
    let seed = "979202789523";
    tx = await contract.drawWinner("activity1", seed);
    receipt = await tx.wait();
    if(receipt.status === 1)
      console.log("已开奖, 交易Hash: %s， 区块Hash: %s，区块高度: %s", 
        receipt.transactionHash, receipt.blockHash, receipt.blockNumber);
    else
      console.log("开奖tx失败: %s", receipt);
    //查询中奖人
    let luckyIndex = await contract.luckyIndexOf("activity1");
    console.log("中奖人索引：%s", luckyIndex);
    let luckyAddress = await contract.luckyOf("activity1");
    console.log("中奖人地址：%s", luckyAddress);
  };
  
  const runMain = async () => {
    try {
      await main();

      process.exit(0);
    } catch (error) {
      console.log(error);
      process.exit(1);
    }
  };
  
  runMain();