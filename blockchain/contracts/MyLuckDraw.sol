// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.17;

import "hardhat/console.sol";


contract Ownable {
    address public owner;

    event OwnerChanged(address indexed oldOwner, address indexed newOwner);

    constructor() {
        owner = msg.sender;
    }

    modifier onlyOwner() {
        require(msg.sender == owner, "Owner account is required");
        _;
    }

    /**
     * @dev Allows the current owner to transfer ownership of the contract to newOwner.
     * @param newOwner The identity to transfer ownership to.
     */
    function changeOwner(address newOwner) public onlyOwner {
        require(newOwner != owner, "New Owner cannot be the current owner");
        require(newOwner != address(0), "New Owner cannot be zero identity");
        address prevOwner = owner;
        owner = newOwner;
        emit OwnerChanged(prevOwner, newOwner);
    }
}

/**
 * 抽奖合约
 * 
*/
contract MyLuckyDraw is Ownable {
    enum Status {
        created,
        active,
        drawing,
        drawn,
        canceled
    }

    //一次活动
    struct LuckyDraw {
        uint256 id;
        string name;
        Status status;
        uint256 startTime;
        uint256 endTime;
        uint256 luckyCount;  //规定本次活动几个中奖人
        address[] participants; //参加报名人的集合
        uint256[] winnerIndexs; //中奖人的报名序号
    }

    //活动数量
    uint256 drawNumber;

    //活动名 -> 活动
    mapping(string => LuckyDraw) allDraws;

    //活动id -> 用户 -> 是否中奖
    mapping(uint256 => mapping(address => bool)) winners;

    //用户address -> 活动id -> 报名号
    mapping(address => mapping(uint256 => uint256)) participants;

    event CreateActivity(uint256 id, string indexed name, uint256 startTime, uint256 endTime, uint256 luckyCount);
    event Participate(string indexed activityName, address user, uint256 participateIndex);

    /*
    constructor() {
        owner = msg.sender;
    }
    */

    modifier onlyActive(string memory activityName, uint256 nowtime) {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(activity.status == Status.active, "Activity is not active!");
        require(nowtime >= activity.startTime, "Activity has not started!");
        require(nowtime < activity.endTime, "Activity has ended!");
        _;
    }

    modifier onlyDrawing(string memory activityName) {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(
            activity.status == Status.drawing,
            "Activity is not at drawing stutas!"
        );
        _;
    }

    /**核心算法
     * 第一个中奖者：keccak256（出块时间+块高+随机种子）得到一个hash，然后将hash转uint256，再对参加人数取模
     * 后续中奖者：keccak256（前一中奖者索引+出块时间+块高+随机种子）得到一个hash，然后将hash转uint256，再对参加人数取模
     */
    function random2(
        uint256 preLuckIndex,
        uint256 number,
        string memory seed,
        uint256 blockNumber,
        uint256 blocTs
    ) private pure returns (uint256) {
        //console.log("seed is", seed);
        if (preLuckIndex == 9999999999) {
            //console.log( keccak256(abi.encodePacked(blocTs, blockNumber, seed)) );
            //console.log("keccak256 hash int value: ", uint256( keccak256(abi.encodePacked(blocTs, blockNumber, seed)) ));
            return uint256( keccak256(abi.encodePacked(blocTs, blockNumber, seed)) ) % number;
        }
        //console.log("keccak256 hash int value: ", uint256( keccak256(abi.encodePacked(preLuckIndex, blocTs, blockNumber, seed)) ));
        return uint256( keccak256(abi.encodePacked(preLuckIndex, blocTs, blockNumber, seed)) ) % number;
    }

    /**
     * 创建活动
     */
    function createActivity(
        string memory name,
        uint256 startTime,
        uint256 endTime,
        uint256 luckyCount
    ) public onlyOwner {
        require(
            bytes(allDraws[name].name).length == 0,
            "Activity is already exist!"
        );
        require(
            startTime < endTime,
            "Activity startTime must be earlier than endTime!"
        );
        require(luckyCount > 0, "Activity luckyCount must be greater than zero!");
        LuckyDraw storage luckyDraw = allDraws[name];
        luckyDraw.id = drawNumber++;
        luckyDraw.name = name;
        luckyDraw.startTime = startTime;
        luckyDraw.endTime = endTime;
        luckyDraw.luckyCount = luckyCount;
        luckyDraw.status = Status.active;
        emit CreateActivity(luckyDraw.id, name, startTime, endTime, luckyCount);
    }

    /**
     * 报名
     */
    function participate(
        string memory activityName,
        uint256 nowtime,
        address user
    ) public onlyOwner onlyActive(activityName, nowtime) returns(uint256) {
        LuckyDraw storage activity = allDraws[activityName];
        require(
            participants[user][activity.id] == 0,
            "User already particepate"
        );
        activity.participants.push(user);
        
        participants[user][activity.id] = activity.participants.length;
        console.log("activity id :", activity.id);
        console.log("baoming No. :", activity.participants.length - 1);
        emit Participate(activityName, user, activity.participants.length - 1);
        return activity.participants.length - 1;
    } //participants里存的是从1开始的序号，返回的是从0开始的。好像为的是participants[user][activity.id] == 0判定


    /**
     * 根据seed执行算法进行抽奖
     */
    function drawWinner(
        string memory activityName,
        string memory seed,
        uint256 nowtime
    ) public onlyOwner returns (uint256[] memory) {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        //require(nowtime >= activity.endTime, "Activity has not ended!");
        require(activity.status == Status.active, "Activity is not active!");
        activity.status = Status.drawn;
        uint256 luckyCount = activity.luckyCount;   //一共多少个中奖
        uint256 totalPartCount = activity.participants.length; //多少个参加者
        //require(luckyCount <= totalPartCount);
        uint256[] storage winnerIndexs = activity.winnerIndexs;
        uint256 preLuckyIndex = 9999999999;
        for (uint256 i = 0; i < luckyCount && i < totalPartCount; i++) {
            console.log("draw seed is:" , seed);
            uint256 luckyIndex = random2(preLuckyIndex, totalPartCount, seed, block.number, block.timestamp);
            console.log("luckIndex is: ", luckyIndex);
            //emit ActivityRandom(activityName, totalPartCount, seed, seed2);
            if (winners[activity.id][activity.participants[luckyIndex]]) {  //如果抽到的人本次活动已经中奖了，往后顺延一个
                luckyIndex = nextLuckyOne(
                    activity.id,
                    activity.participants,
                    luckyIndex,
                    totalPartCount
                );
            }
            preLuckyIndex = luckyIndex;

            winnerIndexs.push(luckyIndex);
            winners[activity.id][activity.participants[luckyIndex]] = true;
        }
        
        console.log("winner is: ");
        for(uint256 i=0; i<winnerIndexs.length; i++) {
            console.log(winnerIndexs[i]);
        }

        return winnerIndexs;
    }

    //多个中奖者的情况，如果抽到了已经中奖的人，那么顺延一个，有小概率出现连号的情况
    function nextLuckyOne(
        uint256 activityId,
        address[] memory activityParts,
        uint256 luckyIndex,
        uint256 totalPartCount
    ) private view returns (uint256) {
        luckyIndex++;
        if (luckyIndex >= totalPartCount) {
            luckyIndex = 0;
        }
        if (winners[activityId][activityParts[luckyIndex]]) {
            luckyIndex = nextLuckyOne(
                activityId,
                activityParts,
                luckyIndex,
                totalPartCount
            );
        }
        return luckyIndex;
    }

    /**
     * 获得中奖者的索引
     */
    function luckyIndexOf(string memory activityName)
        public
        view
        returns (uint256[] memory)
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(activity.status == Status.drawn, "Activity is not finished!");
        return activity.winnerIndexs;
    }

    /**
     * 获得中奖者的address
     */
    function luckyOf(string memory activityName)
        public
        view
        returns (address[] memory)
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(activity.status == Status.drawn, "Activity is not finished!");
        address[] memory luckyAddresses = new address[](activity.luckyCount);
        for (uint256 i = 0; i < activity.winnerIndexs.length; i++) {
            luckyAddresses[i] = activity.participants[activity.winnerIndexs[i]];
        }
        return luckyAddresses;
    }

    /**
     * 获取报名人数
     */
    function participantNumOf(string memory activityName)
        public
        view
        returns (uint256)
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        return activity.participants.length;
    }

    /**
     * 获取活动当前的状态
     */
    function statusOf(string memory activityName) public view returns (Status) {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        return activity.status;
    }

    /**
     * @dev Check whether an identity wins a prize of a lottery.
     * @param activityName The lottery name.
     * @param user The identity of user.
     */
    function ifLuckyOf(string memory activityName, address user)
        public
        view
        returns (bool)
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(activity.status == Status.drawn, "Activity is not finished!");
        return winners[activity.id][user];
    }

    /**
     * @dev Get a user's index of a lottery.
     * @param activityName The lottery name.
     * @param user The identity of user.
     */
    function indexOf(string memory activityName, address user)
        public
        view
        returns (uint256)
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(participants[user][activity.id] != 0, "User not participate!");
        return participants[user][activity.id] - 1;
    }

    /**
     * 取消活动
     */
    function cancelActivity(string memory activityName)
        public
        onlyOwner
    {
        LuckyDraw storage activity = allDraws[activityName];
        require(bytes(activity.name).length != 0, "Activity is not exist!");
        require(activity.status == Status.active, "Activity is not active!");
        activity.status = Status.canceled;
    }
}