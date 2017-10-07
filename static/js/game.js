/* global Phaser RemotePlayer io */


var game = new Phaser.Game(800, 600, Phaser.AUTO, "container", { preload: preload, create: create, update: update, render: render })

function preload () {

  //game.load.image('backdrop', 'assets/pics/cod.jpg');
  //game.load.image('backdrop','assets/virtualjoystick/sky4.png')
  game.load.image('bg','assets/skies/deepblue.png')
  game.load.image('green','assets/demoscene/green_ball.png')
  game.load.image('red_ball','assets/demoscene/ball-tlb.png')
  game.load.image('blue_ball','assets/demoscene/blue_ball.png')


  game.load.image("ball","assets/beball1.png")
  game.load.image('earth', 'assets/light_sand.png')
  game.load.spritesheet('dude', 'assets/dude.png', 64, 64)
  game.load.spritesheet('enemy', 'assets/dude.png', 64, 64)
}

var socket // Socket connection

var land

var player

var enemies

var currentSpeed = 0
var cursors

function create () {
  //socket = io.connect();
 //socket = io.connect("http://127.0.0.1:8000", {port: 8000, transports: ["websocket"]});

  var fontStyle = { font: "bold 16px Arial", fill: "#fff", boundsAlignH: "center", boundsAlignV: "middle" };


  // Resize our game world to be a 2000 x 2000 square
  //game.world.setBounds(-500, -500, 1000, 1000)

  // Our tiled scrolling background
  //land = game.add.tileSprite(0, 0, 800, 600, 'earth')

  land = game.add.sprite(0,0,'bg')
  //game.add.sprite(0, 0, 'bg');
  // table
    var  graphics = game.add.graphics(game.world.centerX,game.world.centerY)
     graphics.lineStyle(0.1, 0xffd900);
     graphics.beginFill(0xffd900, 1);
     graphics.drawEllipse(0, 0, 250, 60);
     graphics.endFill();

  player =game.add.sprite(200, 300, 'ball');
  player.bar = game.add.graphics();
  player.bar.beginFill(0x000000, 0.2);
  player.bar.drawRect(100, 300, 100, 40);
  //  The Text is positioned at 0, 100
  player.text = game.add.text(0, 0, "seat: 0\nname: wade", fontStyle);
  player.text.setShadow(3, 3, 'rgba(0,0,0,0.5)', 2);
  player.text.setTextBounds(100, 300, 100, 40);
  // The base of our player
  //var startX = Math.round(Math.random() * (1000) - 500)
  //var startY = Math.round(Math.random() * (1000) - 500)
  //player = game.add.sprite(startX, startY, 'dude')
  //player.anchor.setTo(0.5, 0.5)
  //player.animations.add('move', [0, 1, 2, 3, 4, 5, 6, 7], 20, true)
  //player.animations.add('stop', [3], 20, true)

  // This will force it to decelerate and limit its speed
  // player.body.drag.setTo(200, 200)
  //game.physics.enable(player, Phaser.Physics.ARCADE);
  //player.body.maxVelocity.setTo(400, 400)
  //player.body.collideWorldBounds = true

  // Create some baddies to waste :)
  enemies = []

  player.bringToTop()

  game.camera.follow(player)
  game.camera.deadzone = new Phaser.Rectangle(150, 150, 500, 300)
  game.camera.focusOnXY(0, 0)

  cursors = game.input.keyboard.createCursorKeys()






  // Start listening for events
  setEventHandlers()
}

var setEventHandlers = function () {
  // Socket connection successful
  //socket.on('connect', onSocketConnected)

  // Socket disconnection
  //socket.on('disconnect', onSocketDisconnect)

  // New player message received
  //socket.on('new player', onNewPlayer)

  // Player move message received
  //socket.on('move player', onMovePlayer)

  // Player removed message received
  //socket.on('remove player', onRemovePlayer)
}

// Socket connected
function onSocketConnected () {
  console.log('Connected to socket server')

  // Reset enemies on reconnect
  enemies.forEach(function (enemy) {
    enemy.player.kill()
  })
  enemies = []

  // Send local player data to the game server
  socket.emit('new player', { x: player.x, y: player.y, angle: player.angle })
}

// Socket disconnected
function onSocketDisconnect () {
  console.log('Disconnected from socket server')
}

// New player
function onNewPlayer (data) {
  console.log('New player connected:', data.id)

  // Avoid possible duplicate players
  var duplicate = playerById(data.id)
  if (duplicate) {
    console.log('Duplicate player!')
    return
  }

  // Add new player to the remote players array
  enemies.push(new RemotePlayer(data.id, game, player, data.x, data.y, data.angle))
}

// Move player
function onMovePlayer (data) {
  var movePlayer = playerById(data.id)

  // Player not found
  if (!movePlayer) {
    console.log('Player not found: ', data.id)
    return
  }

  // Update player position
  movePlayer.player.x = data.x
  movePlayer.player.y = data.y
  movePlayer.player.angle = data.angle
}

// Remove player
function onRemovePlayer (data) {
  var removePlayer = playerById(data.id)

  // Player not found
  if (!removePlayer) {
    console.log('Player not found: ', data.id)
    return
  }

  removePlayer.player.kill()

  // Remove player from array
  enemies.splice(enemies.indexOf(removePlayer), 1)
}

function update () {
  for (var i = 0; i < enemies.length; i++) {
    if (enemies[i].alive) {
      enemies[i].update()
      game.physics.arcade.collide(player, enemies[i].player)
    }
  }

  if (cursors.left.isDown) {
    player.angle -= 4
  } else if (cursors.right.isDown) {
    player.angle += 4
  }

  if (cursors.up.isDown) {
    // The speed we'll travel at
    currentSpeed = 300
  } else {
    if (currentSpeed > 0) {
      currentSpeed -= 4
    }
  }

  //game.physics.arcade.velocityFromRotation(player.rotation, currentSpeed, player.body.velocity)

  if (currentSpeed > 0) {
    player.animations.play('move')
  } else {
    player.animations.play('stop')
  }



  if (game.input.activePointer.isDown) {
    if (game.physics.arcade.distanceToPointer(player) >= 10) {
      currentSpeed = 300

      player.rotation = game.physics.arcade.angleToPointer(player)
    }
  }

  //socket.emit('move player', { x: player.x, y: player.y, angle: player.angle })
}

function render () {

}

// Find player by ID
function playerById (id) {
  for (var i = 0; i < enemies.length; i++) {
    if (enemies[i].player.name === id) {
      return enemies[i]
    }
  }

  return false
}
