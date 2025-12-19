package api

import (
	"encoding/xml"
	"net/http"
)

type SOAPEnvelope struct {
	XMLName xml.Name `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"`
	Body    SOAPBody `xml:"Body"`
}

type SOAPBody struct {
	Content []byte `xml:",innerxml"`
}

func (rt *Router) handleSOAP(w http.ResponseWriter, r *http.Request) {
	var envelope SOAPEnvelope
	if err := xml.NewDecoder(r.Body).Decode(&envelope); err != nil {
		rt.logger.Error("Failed to decode SOAP", err)
		http.Error(w, "Invalid SOAP XML", http.StatusBadRequest)
		return
	}

	// Just log the inner content for now
	rt.logger.Info("Received SOAP content", "length", len(envelope.Body.Content))
	
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`<soap:Envelope xmlns:soap="http://schemas.xmlsoap.org/soap/envelope/"><soap:Body><Response>Ack</Response></soap:Body></soap:Envelope>`))
}
