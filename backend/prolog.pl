sintoma(fiebre).
sintoma(tos).
sintoma(dolor_cabeza).
sintoma(cansancio).
sintoma(dificultad_respirar).
enfermedad(gripe, gripe_comun).
enfermedad(covid19, covid_19).
enfermedad(migrana, migrana).
enfermedad(asma, asma).
enfermedad_sintoma(gripe, fiebre, 0.3).
enfermedad_sintoma(gripe, tos, 0.3).
enfermedad_sintoma(gripe, dolor_cabeza, 0.2).
enfermedad_sintoma(gripe, cansancio, 0.2).
enfermedad_sintoma(covid19, fiebre, 0.3).
enfermedad_sintoma(covid19, tos, 0.3).
enfermedad_sintoma(covid19, dificultad_respirar, 0.4).
enfermedad_sintoma(migrana, dolor_cabeza, 0.7).
enfermedad_sintoma(migrana, cansancio, 0.3).
enfermedad_sintoma(asma, dificultad_respirar, 0.6).
enfermedad_sintoma(asma, tos, 0.4).
cronica(diabetes).
cronica(hipertension).
cronica(asma).
trata(gripe, paracetamol).
trata(migrana, ibuprofeno).
trata(asma, salbutamol).
trata(covid19, paracetamol).
urgencia(Severidades, consulta_medica_inmediata_sugerida) :- member(severo, Severidades), !.
urgencia(Severidades, observacion_recomendada) :- member(moderado, Severidades), !.
urgencia(_, posible_automanejo).
medicamento(paracetamol,paracetamol).
medicamento(ibuprofeno,ibuprofeno).
medicamento(salbutamol,salbutamol).
contraindicacion(ibuprofeno,hipertension).
contraindicacion(salbutamol,diabetes).
