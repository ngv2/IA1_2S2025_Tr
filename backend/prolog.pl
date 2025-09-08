% Hechos
animal(raton).
animal(serpiente).
animal(halcon).
animal(aguila).
animal(leon).

depredador(serpiente, raton).
depredador(halcon, serpiente).
depredador(aguila, halcon).
depredador(leon, aguila).

% Reglas
come(X, Y) :- depredador(X, Y).
come(X, Y) :- depredador(X, Z), come(Z, Y).