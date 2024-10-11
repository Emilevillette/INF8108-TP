# Partie 2.2 README
Les sources pour le credential stealer sont disponible dans le dossier source

## Fonctionnement du Credential stealer

Le fonctionnement du credential stealer est le suivant :
1. Le payload est exécuté sur la machine cible
2. Le stealer va chercher les mots de passe stockés dans le navigateur Chrome
3. Le stealer va chercher les identifiants Wifi stockés sur la machine en utilisant l'API Windows
4. Le stealer envoie les mots de passe et identifiants Wifi à un serveur TCP local sur le port 4444

### Pour d'autres systèmes d'exploitation et navigateurs

Pour d'autres systèmes d'exploitation, tels que MacOS ou Linux, nous pourrions imaginer un payload avec un fichier PDF corrompu, tel que présenté dans la partie 2.1. le reste du code pourrait suivre un modus operandi similaire, bien qu'il faudrait simplement adapter les scripts bat en scripts bash ou autre. Idem pour des navigateurs autres que Chrome, il suffit de changer les chemins d'accès aux fichiers de crédentiels différer d'un système d'explotation, et d'un navigateur à l'autre.

Nous pensons que le proof of concept sur Windows est suffisant pour illustrer le fonctionnement du keylogger et comment il pourrait être adapté à d'autres systèmes d'exploitation.
