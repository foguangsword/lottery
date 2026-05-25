const main = async () => {
    const signers = await hre.ethers.getSigners();
    console.log(signers.length);
    const owner = signers[0];

    const contractFactory = await hre.ethers.getContractFactory('MyLuckyDraw');
    const contract = await contractFactory.deploy();
    await contract.deployed();


    console.log("Contract deployed to: %s, owner: %s", contract.address, owner.address);

    //创建活动
    let create = await contract.createActivity("activity1", new Date().getTime(), new Date().getTime() + 2000, 5);
    await create.wait();
    
    //报名
    for(var i=0; i<2739; i++){
      //const randomPerson = await hre.ethers.getSigners();
      const wallet = await hre.ethers.Wallet.createRandom();
      let baoming = await contract.participate("activity1", new Date().getTime(), wallet.address);
      await baoming.wait();
      console.log("报名人", wallet.address);
      //console.log("报名索引", ethers.BigNumber.from(baoming).toNumber());
    }
    
    
    //模拟报名结束后触发开奖
    let blockTs = 1670061981603; //报名结束存证的时间戳
    let blockNumber = 173637; //报名结束存证的块高
    console.log("报名结束，时间戳：" , blockTs);
    console.log("报名结束，块高：" , blockNumber);
    
    //开奖
    let seed = "979202789523";
    let drawWinner = await contract.drawWinner("activity1", seed, new Date().getTime(), blockNumber, blockTs);
    await drawWinner.wait();
    //console.log("中奖人", drawWinner);

    //几个查询
    

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