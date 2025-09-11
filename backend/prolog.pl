% ===========================
% Hechos: síntomas
% ===========================
% sintoma(Id, Nombre).
sintoma(fiebre, "Fiebre").
sintoma(tos, "Tos persistente").
sintoma(dolor_cabeza, "Dolor de cabeza").
sintoma(cansancio, "Cansancio").
sintoma(dificultad_respirar, "Dificultad para respirar").

% ===========================
% Hechos: enfermedades
% ===========================
% enfermedad(Id, Nombre).
enfermedad(gripe, "Gripe común").
enfermedad(covid19, "COVID-19").
enfermedad(migraña, "Migraña").
enfermedad(asma, "Asma").

% ===========================
% Relación enfermedad - síntomas
% ===========================
% enfermedad_sintoma(Enfermedad, Sintoma, Peso).
enfermedad_sintoma(gripe, fiebre, 0.3).
enfermedad_sintoma(gripe, tos, 0.3).
enfermedad_sintoma(gripe, dolor_cabeza, 0.2).
enfermedad_sintoma(gripe, cansancio, 0.2).

enfermedad_sintoma(covid19, fiebre, 0.3).
enfermedad_sintoma(covid19, tos, 0.3).
enfermedad_sintoma(covid19, dificultad_respirar, 0.4).

enfermedad_sintoma(migraña, dolor_cabeza, 0.7).
enfermedad_sintoma(migraña, cansancio, 0.3).

enfermedad_sintoma(asma, dificultad_respirar, 0.6).
enfermedad_sintoma(asma, tos, 0.4).

% ===========================
% Enfermedades crónicas
% ===========================
cronica(diabetes).
cronica(hipertension).
cronica(asma).

% ===========================
% Medicamentos
% ===========================
% medicamento(Id, Nombre).
medicamento(paracetamol, "Paracetamol").
medicamento(ibuprofeno, "Ibuprofeno").
medicamento(salbutamol, "Salbutamol").

% Reglas de compatibilidad (medicamento recomendado para enfermedad)
% trata(Enfermedad, Medicamento).
trata(gripe, paracetamol).
trata(migraña, ibuprofeno).
trata(asma, salbutamol).
trata(covid19, paracetamol).

% ===========================
% Alergias y contraindicaciones
% ===========================
% contraindicacion(Medicamento, CondicionCronica).
contraindicacion(ibuprofeno, hipertension).
contraindicacion(salbutamol, diabetes).

% ===========================
% Reglas de inferencia
% ===========================
% Calcular afinidad de una enfermedad dado una lista de síntomas del paciente
% afinidad(Enfermedad, ListaSintomas, Afinidad).
afinidad(E, Sints, A) :-
    findall(Peso, (member(S, Sints), enfermedad_sintoma(E, S, Peso)), Pesos),
    sum_list(Pesos, Sum),
    (Sum > 1.0 -> A = 1.0 ; A = Sum).

% Sugerir medicamento válido (evitando alergias y contraindicaciones)
% sugerir_medicamento(Enfermedad, Alergias, Cronicas, Medicamento).
sugerir_medicamento(E, Alergias, Cronicas, M) :-
    trata(E, M),
    \\+ member(M, Alergias),
    \\+ (contraindicacion(M, C), member(C, Cronicas)).

% Nivel de urgencia en base a severidad
% urgencia(ListaSeveridades, Urgencia).
urgencia(Severidades, \"Consulta médica inmediata sugerida\") :- member(severo, Severidades), !.
urgencia(Severidades, \"Observación recomendada\") :- member(moderado, Severidades), !.
urgencia(_, \"Posible automanejo\").

% ===========================
% Ejemplo de consulta:
% ?- afinidad(gripe, [fiebre, tos], A).
% ?- sugerir_medicamento(gripe, [ibuprofeno], [asma], M).
% ?- urgencia([leve, moderado], U).
% ===========================
