require('dotenv').config();
const express = require('express');
const { spawn } = require('child_process');
const { Client, GatewayIntentBits } = require('discord.js');
const { 
    joinVoiceChannel, 
    getVoiceConnection, 
    createAudioPlayer, 
    createAudioResource, 
    AudioPlayerStatus,
    StreamType
} = require('@discordjs/voice');

const app = express();
app.use(express.json());

const client = new Client({
    intents: [
        GatewayIntentBits.Guilds,
        GatewayIntentBits.GuildVoiceStates
    ]
});

const players = new Map();

process.on('uncaughtException', (err) => {
  if (err.code === 'EPIPE' || err.code === 'ECONNRESET') {
    return;
  }
  console.error('Aplikacijska greška:', err);
});
function getOrCreatePlayer(guildId) {
    if (!players.has(guildId)) {
        const player = createAudioPlayer();

        player.on(AudioPlayerStatus.Idle, async () => {
            console.log(`Song on guild: ${guildId}`);
            try {
                await fetch('http://localhost:8080/events/song-ended', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ guildId })
                });
            } catch (err) {
                console.error('Error sending update to Go bot:', err.message);
            }
        });

        player.on('error', error => {
            console.error('AudioPlayer error:', error);
        });

        players.set(guildId, player);
    }
    return players.get(guildId);
}

client.on('clientReady', async () => {
    console.log(`Node Voice Worker turned on: ${client.user.tag}`);

client.guilds.cache.forEach(async (guild) => {
    const me = guild.members.me;
    if (me && me.voice.channelId) {
        console.log(`🧹 Cleaning ghost connection in ${guild.name}...`);
        try {
            const connection = joinVoiceChannel({
                channelId: me.voice.channelId,
                guildId: guild.id,
                adapterCreator: guild.voiceAdapterCreator,
            });
            connection.destroy();
        } catch (err) {
            console.error('Error cleaning up ghost connection:', err);
        }
    }
});
});

client.on('voiceStateUpdate', (oldState, newState) => {
    const guildId = oldState.guild.id || newState.guild.id;
    const connection = getVoiceConnection(guildId);
    if (!connection) return;

    const botChannelId = connection.joinConfig.channelId;
    const guild = client.guilds.cache.get(guildId);
    if (!guild) return;

    const channel = guild.channels.cache.get(botChannelId);
    if (!channel) return;

    const humanMembers = channel.members.filter(member => !member.user.bot);

    if (humanMembers.size === 0) {
        console.log(`🧹 All users left ${guild.name}. Leaving...`);
        
        const player = players.get(guildId);
        if (player) player.stop();

        connection.destroy();
    }
});

app.post('/join', (req, res) => {
    const { guildId, channelId } = req.body;
    if (!guildId || !channelId) return res.status(400).json({ error: 'Missing guildId or channelId' });

    try {
        const guild = client.guilds.cache.get(guildId);
        if (!guild) return res.status(404).json({ error: 'Guild not found' });

        const connection = joinVoiceChannel({
            channelId: channelId,
            guildId: guildId,
            adapterCreator: guild.voiceAdapterCreator,
            selfDeaf: true,
        });

        const player = getOrCreatePlayer(guildId);
        connection.subscribe(player);

        console.log(`🔊 Joined channel ${channelId}`);
        return res.json({ status: 'ok' });
    } catch (err) {
        return res.status(500).json({ error: err.message });
    }
});

app.post('/play', (req, res) => {
    const { guildId, url } = req.body;
    if (!guildId || !url) return res.status(400).json({ error: 'Missing guildId or url' });

    const connection = getVoiceConnection(guildId);
    if (!connection) return res.status(400).json({ error: 'Ajzak is not in a voice channel' });

    try {
        const ytdlp = spawn('yt-dlp', ['-o', '-', '-f', 'bestaudio', url]);
        const ffmpeg = spawn('ffmpeg', [
            '-i', 'pipe:0',
            '-f', 's16le',
            '-ar', '48000',
            '-ac', '2',
            'pipe:1'
        ]);

        ytdlp.stdout.on('error', err => { if (err.code === 'EPIPE') return; });
        ffmpeg.stdin.on('error', err => { if (err.code === 'EPIPE') return; });
        ffmpeg.stdout.on('error', err => { if (err.code === 'EPIPE') return; });

        ytdlp.stdout.pipe(ffmpeg.stdin);

        const resource = createAudioResource(ffmpeg.stdout, {
            inputType: StreamType.Raw
        });

        const player = getOrCreatePlayer(guildId);
        player.play(resource);

        console.log(`▶️ Streaming song for: ${url}`);
        return res.json({ status: 'ok' });
    } catch (err) {
        console.error('Streaming error:', err);
        return res.status(500).json({ error: err.message });
    }
});

app.post('/stop', (req, res) => {
    const { guildId } = req.body;
    const player = players.get(guildId);
    if (player) {
        player.stop(true);
        return res.json({ status: 'ok' });
    }
    return res.status(400).json({ error: 'No active audio player' });
});

app.post('/leave', (req, res) => {
    const { guildId } = req.body;
    const guild = client.guilds.cache.get(guildId);
    if (!guild) {
        return res.status(404).json({ error: 'Guild not found' });
    }
    let connection = getVoiceConnection(guildId);
    if (!connection && guild.members.me?.voice.channelId) {
        connection = joinVoiceChannel({
            channelId: guild.members.me.voice.channelId,
            guildId: guild.id,
            adapterCreator: guild.voiceAdapterCreator,
        });
    }
    if (connection) {
        connection.destroy();
        return res.json({ success: true });
    }
    return res.status(400).json({ error: 'Ajzak is not in a voice channel' });
});

app.listen(3000, () => {
    console.log('Voice Sidecar API listening on http://localhost:3000');
});

const handleShutdown = () => {
    console.log('\nService turning off, disconnecting from voice channels');
    client.guilds.cache.forEach(guild => {
        const connection = getVoiceConnection(guild.id);
        if (connection) connection.destroy();
    });
    process.exit(0);
};

process.on('SIGINT', handleShutdown);
process.on('SIGTERM', handleShutdown);

client.login(process.env.DISCORD_TOKEN);